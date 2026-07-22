#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  printf '%s\n' 'usage: codex-exec-restore-session.sh <prompt-path>' >&2
  exit 2
fi

prompt_path="${1}"

if [[ -z "${MCH_DEFAULT_DIR:-}" ]]; then
  printf '%s\n' 'missing MCH_DEFAULT_DIR' >&2
  exit 1
fi

if [[ "${MCH_DEFAULT_DIR}" == /* || "${MCH_DEFAULT_DIR}" == *".."* ]]; then
  printf 'invalid MCH_DEFAULT_DIR: %s\n' "${MCH_DEFAULT_DIR}" >&2
  exit 1
fi

if [[ -z "${MCH_TEMP_DIR:-}" ]]; then
  printf '%s\n' 'missing MCH_TEMP_DIR' >&2
  exit 1
fi

if [[ "${MCH_TEMP_DIR}" == /* || "${MCH_TEMP_DIR}" == *".."* ]]; then
  printf 'invalid MCH_TEMP_DIR: %s\n' "${MCH_TEMP_DIR}" >&2
  exit 1
fi

if [[ -z "${MCH_REF_UUID:-}" ]]; then
  printf '%s\n' 'missing MCH_REF_UUID' >&2
  exit 1
fi

if [[ ! "${MCH_REF_UUID}" =~ ^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$ ]]; then
  printf 'invalid MCH_REF_UUID: %s\n' "${MCH_REF_UUID}" >&2
  exit 1
fi

if [[ -z "${MCH_STAGE:-}" ]]; then
  printf '%s\n' 'missing MCH_STAGE' >&2
  exit 1
fi
case "${MCH_STAGE}" in
  idea|spec|spec-review|docs|code|pr|code-review|code-docs|merge) ;;
  *) printf 'invalid MCH_STAGE: %s\n' "${MCH_STAGE}" >&2; exit 1 ;;
esac

repo="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${repo}" ]]; then
  printf '%s\n' 'cannot extract repo root' >&2
  exit 1
fi

if [[ "${prompt_path}" != /* ]]; then
  prompt_path="${PWD}/${prompt_path}"
fi

if [[ ! -f "${prompt_path}" ]]; then
  printf 'missing prompt file: %s\n' "${prompt_path}" >&2
  exit 1
fi

flow_temp_root="${repo}/${MCH_TEMP_DIR}"
if [[ ! -d "${flow_temp_root}" ]]; then
  printf 'missing Flow temp root: %s\n' "${MCH_TEMP_DIR}" >&2
  exit 1
fi

temp_dir="${flow_temp_root}/${MCH_REF_UUID}/${MCH_STAGE}"

for resource in input.md output.md; do
  if [[ ! -f "${temp_dir}/${resource}" ]]; then
    printf 'missing artifact resource: %s/%s/%s/%s\n' "${MCH_TEMP_DIR}" "${MCH_REF_UUID}" "${MCH_STAGE}" "${resource}" >&2
    exit 1
  fi
done

prompt="$(
  sed \
    -e "s|/stg-tmp-dir/|${temp_dir}/|g" \
    -e "s|/def-dir/|${MCH_DEFAULT_DIR%/}/|g" \
    "${prompt_path}"
)"

rm -f \
  "${temp_dir}/agent-output.md" \
  "${temp_dir}/events.jsonl" \
  "${temp_dir}/error.log" \
  "${temp_dir}/session-id"

render_codex_progress_line() {
  local status_text="${1}"
  local elapsed_us="${2}"
  local progress_dots="${3}"
  local finish_line="${4}"
  local pulse_frames=('·' '•' '●' '•')
  local pulse_count="${#pulse_frames[@]}"
  local pulse_frame_duration_us=250000
  local pulse_index=$(((elapsed_us / pulse_frame_duration_us) % pulse_count))
  local pulse="${pulse_frames[pulse_index]}"
  local elapsed=$((elapsed_us / 1000000))
  local minutes=$((elapsed / 60))
  local seconds=$((elapsed % 60))

  if [[ -t 1 ]]; then
    printf '\r\033[2K'
  fi

  printf '%s - %02d:%02d - %s%s' \
    "${status_text}" \
    "${minutes}" \
    "${seconds}" \
    "${progress_dots}" \
    "${pulse}"

  if [[ "${finish_line}" == true ]]; then
    printf '\n'
  fi
}

print_codex_progress() {
  local restored_session_id="${1}"
  local event
  local read_status
  local read_timeout
  local started_session_id
  local status_text=""
  local progress_dots=""
  local progress_started_at_us="${EPOCHREALTIME/./}"
  local progress_now_us
  local elapsed_us
  local pulse_frame_duration_us=250000
  local tick_remaining_us
  local is_terminal=false

  if [[ -t 1 ]]; then
    is_terminal=true
  fi

  if [[ -n "${restored_session_id}" ]]; then
    status_text="${restored_session_id} restored"
    if [[ "${is_terminal}" == true ]]; then
      render_codex_progress_line "${status_text}" 0 "${progress_dots}" false
    fi
  fi

  while true; do
    progress_now_us="${EPOCHREALTIME/./}"
    elapsed_us=$((progress_now_us - progress_started_at_us))
    tick_remaining_us=$((pulse_frame_duration_us - (elapsed_us % pulse_frame_duration_us)))
    printf -v read_timeout '%d.%06d' \
      "$((tick_remaining_us / 1000000))" \
      "$((tick_remaining_us % 1000000))"

    event=""
    read_status=0
    IFS= read -r -t "${read_timeout}" event || read_status=$?

    if ((read_status == 0)) || [[ -n "${event}" ]]; then
      progress_dots+='.'

      if [[ -z "${status_text}" ]]; then
        started_session_id="$(
          jq -r '
            if .type == "thread.started" then
              .thread_id // .session_id // .session.id // .id // empty
            else
              empty
            end
          ' <<< "${event}" 2>/dev/null || true
        )"
        if [[ -n "${started_session_id}" ]]; then
          status_text="${started_session_id} started"
        fi
      fi
    elif ((read_status <= 128)); then
      break
    fi

    if [[ "${is_terminal}" == true && -n "${status_text}" ]]; then
      progress_now_us="${EPOCHREALTIME/./}"
      elapsed_us=$((progress_now_us - progress_started_at_us))
      render_codex_progress_line "${status_text}" "${elapsed_us}" "${progress_dots}" false
    fi
  done

  if [[ -n "${status_text}" ]]; then
    progress_now_us="${EPOCHREALTIME/./}"
    elapsed_us=$((progress_now_us - progress_started_at_us))
    render_codex_progress_line "${status_text}" "${elapsed_us}" "${progress_dots}" true
  fi
}

codex_command=(codex exec -C "${repo}" --json -o "${temp_dir}/agent-output.md" "${prompt}")

set +e
"${codex_command[@]}" 2> "${temp_dir}/error.log" \
  | tee "${temp_dir}/events.jsonl" \
  | print_codex_progress ""
pipeline_status=("${PIPESTATUS[@]}")
set -e

for status in "${pipeline_status[@]}"; do
  if ((status != 0)); then
    exit "${status}"
  fi
done

session_id="$(jq -rsr 'map(select(.type=="thread.started") | (.thread_id // .session_id // .session.id // .id // empty)) | first // empty' "${temp_dir}/events.jsonl")"
if [[ -z "${session_id}" ]]; then
  printf '%s\n' 'missing session-id' >&2
  exit 1
fi
printf '%s\n' "${session_id}" > "${temp_dir}/session-id"

if [[ ! -f "${temp_dir}/agent-output.md" ]]; then
  printf '%s\n' 'missing agent output' >&2
  exit 1
fi
