#!/bin/sh -eu
[ -n "${GOPATH:-}" ] && export "PATH=${GOPATH}/bin:${PATH}"

# Don't translate lxc output for parsing in it in tests.
export LC_ALL="C"

# Force UTC for consistency
export TZ="UTC"

export DEBUG=""
if [ -n "${LXD_VERBOSE:-}" ]; then
  DEBUG="--verbose"
fi

if [ -n "${LXD_DEBUG:-}" ]; then
  DEBUG="--debug"
fi

if [ -n "${DEBUG:-}" ]; then
  set -x
fi

if [ -z "${LXD_BACKEND:-}" ]; then
    LXD_BACKEND="dir"
fi

# shellcheck disable=SC2034
LXD_NETNS=""

import_subdir_files() {
    test "$1"
    # shellcheck disable=SC2039,3043
    local file
    for file in "$1"/*.sh; do
        # shellcheck disable=SC1090
        . "$file"
    done
}

import_subdir_files includes

echo "==> Checking for dependencies"
check_dependencies lxd lxc curl dnsmasq jq git xgettext sqlite3 msgmerge msgfmt shuf setfacl socat dig

if [ "${USER:-'root'}" != "root" ]; then
  echo "The testsuite must be run as root." >&2
  exit 1
fi

if [ -n "${LXD_LOGS:-}" ] && [ ! -d "${LXD_LOGS}" ]; then
  echo "Your LXD_LOGS path doesn't exist: ${LXD_LOGS}"
  exit 1
fi

echo "==> Available storage backends: $(available_storage_backends | sort)"
if [ "$LXD_BACKEND" != "random" ] && ! storage_backend_available "$LXD_BACKEND"; then
  if [ "${LXD_BACKEND}" = "ceph" ] && [ -z "${LXD_CEPH_CLUSTER:-}" ]; then
    echo "Ceph storage backend requires that \"LXD_CEPH_CLUSTER\" be set."
    exit 1
  fi
  echo "Storage backend \"$LXD_BACKEND\" is not available"
  exit 1
fi
echo "==> Using storage backend ${LXD_BACKEND}"

import_storage_backends

cleanup() {
  # Allow for failures and stop tracing everything
  set +ex
  DEBUG=

  # Allow for inspection
  if [ -n "${LXD_INSPECT:-}" ]; then
    if [ "${TEST_RESULT}" != "success" ]; then
      echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION}"
    fi
    echo "==> Test result: ${TEST_RESULT}"

    # shellcheck disable=SC2086
    printf "To poke around, use:\\n LXD_DIR=%s LXD_CONF=%s sudo -E %s/bin/lxc COMMAND\\n" "${LXD_DIR}" "${LXD_CONF}" ${GOPATH:-}
    echo "Tests Completed (${TEST_RESULT}): hit enter to continue"
    read -r _
  fi

  if [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "==> Skipping cleanup (GitHub Action runner detected)"
  else
    echo "==> Cleaning up"

    umount -l "${TEST_DIR}/dev"
    kill_external_auth_daemon "$TEST_DIR"
    cleanup_lxds "$TEST_DIR"
  fi

  echo ""
  echo ""
  if [ "${TEST_RESULT}" != "success" ]; then
    echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION}"
  fi
  echo "==> Test result: ${TEST_RESULT}"
}

# Must be set before cleanup()
TEST_CURRENT=setup
TEST_CURRENT_DESCRIPTION=setup
# shellcheck disable=SC2034
TEST_RESULT=failure

trap cleanup EXIT HUP INT TERM

# Import all the testsuites
import_subdir_files suites

# Setup test directory
TEST_DIR=$(mktemp -d -p "$(pwd)" tmp.XXX)
chmod +x "${TEST_DIR}"

if [ -n "${LXD_TMPFS:-}" ]; then
  mount -t tmpfs tmpfs "${TEST_DIR}" -o mode=0751 -o size=6G
fi

mkdir -p "${TEST_DIR}/dev"
mount -t tmpfs none "${TEST_DIR}"/dev
export LXD_DEVMONITOR_DIR="${TEST_DIR}/dev"

LXD_CONF=$(mktemp -d -p "${TEST_DIR}" XXX)
export LXD_CONF

LXD_DIR=$(mktemp -d -p "${TEST_DIR}" XXX)
export LXD_DIR
chmod +x "${LXD_DIR}"
spawn_lxd "${LXD_DIR}" true
LXD_ADDR=$(cat "${LXD_DIR}/lxd.addr")
export LXD_ADDR

start_external_auth_daemon "${LXD_DIR}"

run_test() {
  TEST_CURRENT=${1}
  TEST_CURRENT_DESCRIPTION=${2:-${1}}
  TEST_UNMET_REQUIREMENT=""

  echo "==> TEST BEGIN: ${TEST_CURRENT_DESCRIPTION}"
  START_TIME=$(date +%s)

  # shellcheck disable=SC2039,3043
  local skip=false

  # Skip test if requested.
  if [ -n "${LXD_SKIP_TESTS:-}" ]; then
    for testName in ${LXD_SKIP_TESTS}; do
      if [ "test_${testName}" = "${TEST_CURRENT}" ]; then
          echo "==> SKIP: ${TEST_CURRENT} as specified in LXD_SKIP_TESTS"
          skip=true
          break
      fi
    done
  fi

  if [ "${skip}" = false ]; then
    # Run test.
    ${TEST_CURRENT}

    # Check whether test was skipped due to unmet requirements, and if so check if the test is required and fail.
    if [ -n "${TEST_UNMET_REQUIREMENT}" ]; then
      if [ -n "${LXD_REQUIRED_TESTS:-}" ]; then
        for testName in ${LXD_REQUIRED_TESTS}; do
          if [ "test_${testName}" = "${TEST_CURRENT}" ]; then
              echo "==> REQUIRED: ${TEST_CURRENT} ${TEST_UNMET_REQUIREMENT}"
              false
              return
          fi
        done
      else
        # Skip test if its requirements are not met and is not specified in required tests.
        echo "==> SKIP: ${TEST_CURRENT} ${TEST_UNMET_REQUIREMENT}"
      fi
    fi
  fi

  END_TIME=$(date +%s)

  echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION} ($((END_TIME-START_TIME))s)"
}

# allow for running a specific set of tests
if [ "$#" -gt 0 ] && [ "$1" != "all" ] && [ "$1" != "cluster" ] && [ "$1" != "standalone" ]; then
  run_test "test_${1}"
  # shellcheck disable=SC2034
  TEST_RESULT=success
  exit
fi

if [ "${1:-"all"}" != "standalone" ]; then
    run_test test_clustering_enable "clustering enable"
    run_test test_clustering_membership "clustering membership"
    run_test test_clustering_containers "clustering containers"
    run_test test_clustering_storage "clustering storage"
    run_test test_clustering_storage_single_node "clustering storage single node"
    run_test test_clustering_network "clustering network"
    run_test test_clustering_publish "clustering publish"
    run_test test_clustering_profiles "clustering profiles"
    run_test test_clustering_join_api "clustering join api"
    run_test test_clustering_shutdown_nodes "clustering shutdown"
    run_test test_clustering_projects "clustering projects"
    run_test test_clustering_update_cert "clustering update cert"
    run_test test_clustering_update_cert_reversion "clustering update cert reversion"
    run_test test_clustering_address "clustering address"
    run_test test_clustering_image_replication "clustering image replication"
    run_test test_clustering_dns "clustering DNS"
    run_test test_clustering_recover "clustering recovery"
    run_test test_clustering_handover "clustering handover"
    run_test test_clustering_rebalance "clustering rebalance"
    run_test test_clustering_remove_raft_node "clustering remove raft node"
    run_test test_clustering_failure_domains "clustering failure domains"
    run_test test_clustering_image_refresh "clustering image refresh"
    run_test test_clustering_evacuation "clustering evacuation"
    run_test test_clustering_instance_placement_scriptlet "clustering instance placement scriptlet"
    run_test test_clustering_move "clustering move"
    run_test test_clustering_edit_configuration "clustering config edit"
    run_test test_clustering_remove_members "clustering config remove members"
    run_test test_clustering_autotarget "clustering autotarget member"
    # run_test test_clustering_upgrade "clustering upgrade"
    run_test test_clustering_groups "clustering groups"
    run_test test_clustering_events "clustering events"
    run_test test_clustering_uuid "clustering uuid"
fi


# shellcheck disable=SC2034
TEST_RESULT=success
