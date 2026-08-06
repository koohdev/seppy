---
name: laws-of-ux
description: Apply psychological principles to digital interface and product design (Jakob's Law, Fitts's Law, Miller's Law, Hick's Law, Postel's Law, Peak-End Rule, Aesthetic-Usability Effect, Von Restorff Effect, Tesler's Law, Doherty Threshold). Use when evaluating, auditing, or designing user interfaces, visual hierarchy, interaction flows, accessibility, or UX strategy.
---

# Laws of UX Skill

Use this skill to guide design decisions, conduct UX audits, justify UI choices to stakeholders, and craft human-centered digital experiences using psychological principles.

---

## Quick Reference: Core UX Laws

| Principle | Core Concept | Primary UX Application |
| :--- | :--- | :--- |
| **Jakob's Law** | Users prefer sites/apps to work like those they already know. | Use familiar design patterns, navigation conventions, and standard UI components to reduce cognitive load. |
| **Fitts's Law** | Touch/click time is a function of target distance and target size. | Make interactive elements (CTA buttons, touch targets) larger and place them close to users' natural focus/thumb zones. |
| **Miller's Law** | Working memory capacity is limited to $7 \pm 2$ items. | Chunk complex content and menu items into small, digestible groups. |
| **Hick's Law** | Decision time increases with the number and complexity of choices. | Minimize options during key workflows (e.g., onboarding, checkout); highlight recommended defaults. |
| **Postel's Law** | Be conservative in what you do, liberal in what you accept. | Accept flexible user input formats (dates, phone numbers) while producing clean, standardized output. |
| **Peak–End Rule** | Experiences are judged by their peak moment and end outcome. | Optimize critical conversion moments, delightful micro-interactions, and positive exit/confirmation states. |
| **Aesthetic–Usability Effect** | Aesthetically pleasing design is perceived as more usable. | Invest in visual polish, harmonious color palettes, typography, and fluid micro-animations to enhance user trust and perceived usability. |
| **Von Restorff Effect** | Distinctive items are most likely to be remembered. | Visually differentiate call-to-action buttons, primary pricing tiers, or high-priority notifications. |
| **Tesler's Law** | Every system has an inherent level of complexity that cannot be reduced. | Handle necessary complexity on the system side rather than pushing it onto the user. |
| **Doherty Threshold** | Computer & user interaction speed should be $< 400\text{ ms}$ for continuous flow. | Use skeleton loaders, optimistic UI updates, and responsive feedback to ensure sub-400ms visual response times. |

---

## Design & Evaluation Workflow

When auditing or designing an interface, execute this systematic process:

- [ ] **Step 1: Mental Model Alignment (Jakob's & Postel's)**
  - Verify layout adheres to familiar platform conventions (e.g., header search, standard bottom navigation on mobile).
  - Ensure input fields accept flexible user entries without throwing brittle validation errors.

- [ ] **Step 2: Choice & Memory Optimization (Hick's & Miller's)**
  - Count active options per screen. Break down multi-step forms into progressive disclosure flows.
  - Chunk long lists into groups of 5–9 items max.

- [ ] **Step 3: Interaction & Ergonomics (Fitts's & Doherty)**
  - Ensure touch targets are at least $48\times 48\text{px}$ (mobile) or easily targetable.
  - Ensure visual response/feedback occurs within 400ms (use progress indicators or skeleton states for network latency).

- [ ] **Step 4: Visual Polish & Attention Guidance (Von Restorff & Aesthetic-Usability)**
  - Establish clear visual hierarchy using distinct styling for primary actions.
  - Apply cohesive color palettes, typography rules, and micro-interactions.

- [ ] **Step 5: Emotional & Complexity Management (Peak-End Rule & Tesler's)**
  - Identify the peak emotional moment of the flow and ensure delight or clarity.
  - Offload technical/data transformation tasks from the user to system logic.

---

## Output Template for UX Audits

When generating a Laws of UX evaluation report, use the following structure:

```markdown
# UX Audit Report: [Page / Feature Name]

## Executive Summary
[Brief summary of overall usability, strengths, and primary friction points]

## Key Findings & UX Law Mapping

### 1. [Law Name - e.g., Fitts's Law]
- **Observation:** [What was found in the interface]
- **Psychological Impact:** [Why this impacts the user's mental model or cognitive load]
- **Recommendation:** [Specific actionable UI change]

### 2. [Law Name - e.g., Hick's Law]
- **Observation:** [What was found in the interface]
- **Psychological Impact:** [Why this impacts the user's mental model or cognitive load]
- **Recommendation:** [Specific actionable UI change]

## Priority Action Items
1. **High Priority:** [Immediate fix]
2. **Medium Priority:** [Usability polish]
3. **Low Priority:** [Nice-to-have visual alignment]
```

---

## Gotchas

- **Do not blindly simplify (Tesler's Law):** Over-simplifying UI can sometimes obscure necessary controls or transfer complexity to another step in an inconvenient way.
- **Do not rely on aesthetics alone (Aesthetic-Usability Effect):** A visually stunning application can mask usability issues during initial testing, but user frustration will eventually emerge if underlying functionality breaks.
- **Avoid over-differentiation (Von Restorff Effect):** If every button or card is highlighted as "featured" or "urgent", nothing stands out and cognitive load spikes.
