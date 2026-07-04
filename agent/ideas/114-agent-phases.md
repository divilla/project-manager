# Agent Phases

Input:  /tmp/mch/input.md
Output: /tmp/mch/output.md
Prompt: Done.

Skill: prompt/temp

State:
- name
- help
- entry_script
- prompt
- exit_script
- next_state
- done

AppState:
- idle
- entry_script
- live_agent -> approve_prompt - yes/no
- approve_prompt_yes -> exit_script -> done
- shell_agent
- exit
- done

flow: [<state>]
stop: [<state>]


Implement agent_phase:
- idea - 
- spec
- ready
- docs
- code
- vibe
- pr
- review
- fix
- done
- merge
- stage
- master
