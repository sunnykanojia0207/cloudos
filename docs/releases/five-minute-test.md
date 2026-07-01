# CloudOS v0.6 — Five-Minute Test

## Purpose

Validate that a developer who has never seen CloudOS can install it, deploy
an application, understand what happened, and recommend it to another developer.

## Participants

| # | Name | Background | OS | Date |
|:-:| :--- | :--------- | :- | :--- |
| 1 | | | | |
| 2 | | | | |
| 3 | | | | |

> Target: 3 external developers with no prior CloudOS experience.

---

## Setup

1. Provide the participant with:
   - The URL: `https://github.com/cloudos/cloudos`
   - The command: `curl -fsSL https://cloudos.io/install.sh | sh`
   - The example repo URL: `https://github.com/cloudos-examples/go-api`

2. Do NOT provide any additional instructions.
   The participant should rely entirely on the README and documentation.

3. Start timing when the participant opens the README.

---

## Measurements

### Installation

| Metric | P1 | P2 | P3 | Target |
| :----- | :-: | :-: | :-: | :----: |
| Time to install CloudOS | | | | < 2 min |
| Time to run `cloudosctl doctor` | | | | < 1 min |
| Time to resolve doctor issues | | | | < 2 min |

**Total installation time target: < 3 minutes**

### First Deployment

| Metric | P1 | P2 | P3 | Target |
| :----- | :-: | :-: | :-: | :----: |
| Time to find deploy command | | | | < 30s |
| Time to execute first deploy | | | | < 1 min |
| Time to first successful URL | | | | < 2 min |

**Total first deployment target: < 2 minutes**

### Understanding

| Question | P1 | P2 | P3 |
| :------- | :-: | :-: | :-: |
| Could you find the application URL? | | | |
| Could you verify the app is healthy? | | | |
| Could you view the deployment timeline? | | | |
| Could you stream the logs? | | | |
| Could you list all applications? | | | |
| Could you understand what each CLI command does? | | | |

### Recommendation

| Question | P1 | P2 | P3 |
| :------- | :-: | :-: | :-: |
| Would you recommend CloudOS to another developer? | | | |
| What was the most confusing part? | | | |
| What was missing from documentation? | | | |
| What would make you try CloudOS for a real project? | | | |

---

## Observations

### Participant 1

| Aspect | Observation |
| :----- | :---------- |
| Installation | |
| First deployment | |
| Most confusing | |
| Questions asked | |
| Failures encountered | |
| Suggested improvements | |

### Participant 2

| Aspect | Observation |
| :----- | :---------- |
| Installation | |
| First deployment | |
| Most confusing | |
| Questions asked | |
| Failures encountered | |
| Suggested improvements | |

### Participant 3

| Aspect | Observation |
| :----- | :---------- |
| Installation | |
| First deployment | |
| Most confusing | |
| Questions asked | |
| Failures encountered | |
| Suggested improvements | |

---

## Summary

| Metric | Average | Best | Worst | Target | Met? |
| :----- | :-----: | :--: | :---: | :----: | :--: |
| Installation time | | | | < 3 min | |
| First deployment time | | | | < 2 min | |
| End-to-end time | | | | < 5 min | |

---

## Action Items

| # | Issue | Fix | Status |
|:-:| :---- | :-- | :----- |
| | | | |
| | | | |
| | | | |

---

## Sign-Off

| Role | Name | Date |
| :--- | :--- | :--- |
| Test Coordinator | | |
| All feedback addressed | | |
| Documentation updated | | |
| Release ready | | |
