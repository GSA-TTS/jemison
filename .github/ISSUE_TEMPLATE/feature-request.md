name: Feature Request 💡
description:
  Suggest a new idea for the design system.
title: 'Search.gov - Feature: [YOUR TITLE]'
labels: ['Type: Feature Request','Status: Triage']
body:
  - type: markdown
    attributes:
      value: '## Feature Request 💡'
  - type: textarea
    id: problem
    attributes:
      label: Is your feature request related to a problem? Please describe.
      description: "Provide a clear and concise description of what the problem is. Ex. I'm always frustrated when [...]"
    validations:
      required: true
  - type: textarea
    id: solution
    attributes:
      label: "Describe the solution you'd like"
      description: "Provide a clear and concise description of what you want to happen."
    validations:
      required: true
  - type: textarea
    id: alternatives
    attributes:
      label: "Describe alternatives you've considered"
      description: "Provide a clear and concise description of any alternative solutions or features you've considered."
    validations:
      required: false
  - type: textarea
    id: context
    attributes:
      label: Additional context
      description: "Add any other context or screenshots about the feature request."
    validations:
      required: false
  - type: checkboxes
    id: terms
    attributes:
      label: Code of Conduct
      description: Please confirm the following
      options:
        - label:
            I agree to abide by the [Digital.gov Community Guidelines](https://digital.gov/communities/community-guidelines/) and the [TTS Code of Conduct](https://handbook.tts.gsa.gov/code-of-conduct/). Respect your peers, use plain language, and be patient.
          required: true
        - label:
            I checked the [current
            issues](https://github.com/GSA-TTS/FAC/issues?q=is%3Aissue+is%3Aopen+label%3A%22Type%3A+Feature+Request%22) for
            duplicate feature requests.
          required: true