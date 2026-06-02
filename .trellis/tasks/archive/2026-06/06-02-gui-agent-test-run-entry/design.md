# Design

## Frontend

Add a `Test run` button to the existing agent form actions. It is visible only when editing an existing agent.

On click:

1. Set `#chat-agent` to `state.editingAgent`.
2. Set `#chat-new-session` checked.
3. Fill `#chat-input` with a simple test prompt.
4. Switch to the Chat panel.
5. Focus the chat input.

## Backend

No backend changes. The existing chat form submits to `/message` with the selected agent.

## Compatibility

No API or persistence changes.
