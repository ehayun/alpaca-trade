# Lessons

- **Unclear request = ask, never guess a value.** "The side thresholds are too wide" was read as "lower the
  scanner limit", and 2% → 1% was picked without asking. The user never gave a number. When a request can mean
  two things (scanner limit vs. side panel width) or needs a value the user did not give, ask before editing.
- **Read shared state before a loop that changes it.** "Select all" read `allBox.checked` inside the loop, but each
  row's change handler called `syncSelection()`, which reset `allBox` after the first row. Only row 1 got ticked.
  Store the value in a local first, and trace the handlers that a dispatched event triggers.
- **Every change increases the version.** Bump `?ver=v-X.Y.Z` in `resource/template/index.html` (last number +1)
  in every change set, not only when `app.css` changes.
- **Research = read the docs.** When asked to write guidance for a framework (GoFrame), fetch its official docs. Don't install its CLI or scaffold a demo. Install tools only during a real build, after the user agrees.
