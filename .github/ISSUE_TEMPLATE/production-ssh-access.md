name: Log of temporarily enabling production ssh access
description: Used to record when and why we enabled ssh access to production.
title: "[Production ssh access]: "
labels: ["prodssh"]
projects: ["GSA-TTS/60"]
body:
  - type: markdown
    attributes:
      value: |
        Use this only **after** production ssh access has been disabled again; do not create a ticket before that has happened.

        Assign the people who were involved in the ssh access to this ticket.

        Close the ticket after it’s created; this is for logging activity, not tracking work yet to be done.
  - type: input
    id: prodssh-start
    attributes:
      label: Start (UTC)
      description: UTC timestamp for when production ssh was enabled
      placeholder: 2023-10-11 19:14
    validations:
      required: true
  - type: input
    id: prodssh-end
    attributes:
      label: End (UTC)
      description: UTC timestamp for when production ssh was disabled
      placeholder: 2023-10-11 19:19
    validations:
      required: true
  - type: textarea
    id: explanation
    attributes:
      label: Explanation
      description: Outline why access was enabled and what was done. Link to an issue if possible.
      placeholder:
    validations:
      required: true