# CLI Options, Caching and Agent

- CLI needs to pull all the options and epics from backend on app start
- CLI needs to write options and epics data to:
  /tmp/mch/phase-options.md
  /tmp/mch/agent-phase-options.md
  /tmp/mch/type-options.md
  /tmp/mch/epic-options.md
- Introduce view vw_epic with name and title xxx (#21)
- Instruct agent to load options in skills

- When ref is null display blank instead of id:xxx
