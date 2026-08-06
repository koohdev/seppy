<!--
SYNC IMPACT REPORT
==================
Version change: [TEMPLATE] → 1.0.0 (initial ratification from design system)
Bump type: MAJOR — first concrete fill of the template; all placeholders replaced.

Principles added (6 new):
  I.  Warm Editorial Identity
  II. Typographic Discipline
  III. Colour Restraint & Coral Scarcity
  IV. Surface Rhythm & Elevation
  V.  Anonymity, Permanence & Safety
  VI. Accessibility & Responsiveness

Sections added:
  - Product Constraints
  - Design Governance

Templates reviewed:
  - .specify/templates/plan-template.md   ✅ compatible — Constitution Check gate references this document
  - .specify/templates/spec-template.md   ✅ compatible — no principle-breaking references
  - .specify/templates/tasks-template.md  ✅ compatible — no principle-breaking references

Follow-up TODOs:
  - RATIFICATION_DATE set to 2026-07-27 (today, first ratification)
  - Animation / transition timings intentionally deferred (documented in DESIGN.md Known Gaps)
  - Copernicus and StyreneB are licensed fonts; open-source substitutes are documented in Principle II
-->

# What Made You Smile Today? — Constitution

## Core Principles

### I. Warm Editorial Identity

Every surface MUST be anchored on a warm cream canvas, never pure white and never cool gray.
The platform's defining visual claim — shared with its design inspiration — is warmth over
neutrality. Cool grays, cool blues, and pure white are explicitly out of brand.

Headlines MUST use a slab-serif display face (Cormorant Garamond or EB Garamond as open-source
substitutes for Copernicus / Tiempos Headline) at weight 400 with negative letter-spacing.
Body copy MUST use a humanist sans-serif face (Inter as primary substitute for StyreneB).
The serif/sans split is the editorial voice; switching either to a neutral face breaks it.

Whitespace between major page sections MUST be generous (96 px rhythm). Cards MUST breathe
with internal padding of no less than 24 px. The pacing must feel like a literary publication,
not a dense product dashboard.

**Rationale**: The platform shares DNA with helloimsorry.co — a soft, human, emotionally safe
space. The warm editorial identity signals that this is not a social media feed, not a news
site, not an AI tool. The warmth earns emotional openness from users.

### II. Typographic Discipline

Display sizes MUST use weight 400 (regular), never bold. Negative letter-spacing on display
headlines is non-negotiable: the range is −0.3 px at the smallest display size to −1.5 px at
the largest. Copernicus / Tiempos Headline without negative tracking reads as off-brand.

Body paragraphs MUST use weight 400. Labels, navigation items, badges, and emphasized phrases
MUST use weight 500. No other weights are in scope.

The display/body font split MUST NOT be collapsed into a single typeface. Using Inter or any
humanist sans for display headlines removes the literary editorial character that separates this
platform from generic web products.

Monospace type (JetBrains Mono or system monospace equivalent) is reserved for code contexts
only. It MUST NOT appear in UI labels, navigation, or body copy.

**Rationale**: Typography is the primary carrier of brand voice. The serif headline + humanist
sans body pairing creates a considered, human tone that invites emotional sharing.

### III. Colour Restraint and Coral Scarcity

The colour system MUST operate within exactly three surface tones:
- Warm cream canvas as the default page floor
- Slightly darker cream for content and feature cards
- Dark warm navy for product surfaces, pre-footer CTAs, and the footer itself

A fourth surface tone (purple sections, green bands, blue cards) MUST NOT be introduced.

Coral is the signature accent colour. It MUST be used on primary CTA buttons and on full-bleed
callout card backgrounds. It MUST NOT be painted across incidental UI elements, decorative
borders, or secondary highlights. Coral scarcity is what gives the coral voltage.

Mood theming for submission cards MUST stay within the warm/cool pairing defined in the product
spec — warm amber tones for Smile, cool indigo tones for Sad — and MUST NOT introduce new hues
outside the established palette.

**Rationale**: The cream-coral-navy trinity is the pacing mechanism. Introducing a fourth tone
dilutes contrast and removes the visual rhythm that guides users through the page.

### IV. Surface Rhythm and Elevation

Depth MUST come from colour-block contrast first, shadows last. Shadows are permitted only as
subtle hover-elevation hints at very low opacity and MUST NOT be the primary elevation signal.

The surface alternation pattern across page sections MUST follow: cream → cream-card → dark
mockup → cream → coral callout → dark footer. No two consecutive bands of the same surface
mode are permitted.

Border radius MUST follow the hierarchical scale: 8 px for buttons and inputs, 12 px for
content and product cards, 16 px for hero containers, pill radius for badges. Custom or
intermediate radii MUST be justified.

Buttons MUST have only two defined states: default and active/pressed (darkened). Hover states
beyond what the design system already encodes MUST NOT be added.

**Rationale**: Consistent surface rhythm creates predictable visual pacing. The rule exists to
prevent the incremental accumulation of shadow layers, glow effects, and one-off elevations
that erode design coherence over time.

### V. Anonymity, Permanence, and Safety

Every submission is anonymous. No authentication, no user identity, and no session-linked
identity MUST be stored or associated with a submission. IP addresses, if used for rate
limiting, MUST be hashed before any form of storage; plaintext IP addresses MUST NOT be
persisted.

Submissions are permanent. No delete endpoint, no edit endpoint, and no user-facing deletion
interface MUST be built for v1. This is a product principle, not a technical limitation.

All user-submitted text MUST pass a bilingual profanity screening pipeline before storage.
If any external screening service is unavailable, the system MUST apply a conservative reject
policy — the submission is blocked and the user is prompted to retry. Optimistic pass-through
on API failure is prohibited.

Rate limiting MUST be enforced at the server level. The rate limit threshold is 5 submissions
per IP address per 10-minute window. Client-side rate limit enforcement is UX only and MUST
NOT be the sole enforcement layer.

**Rationale**: Anonymity and permanence are the product's emotional contract with the user.
Safety enforcement (profanity, rate limiting) protects the shared space from abuse without
requiring user accounts.

### VI. Accessibility and Responsiveness

The platform MUST be usable on screens 375 px wide and above. Layout MUST adapt using column
reduction (single-column on mobile) rather than scaling cards down or truncating content.

Primary interactive elements (buttons, inputs, tappable cards) MUST meet a minimum touch
target size of 40 × 40 px. Touch targets below 40 px require explicit justification.

Code blocks and preformatted content inside dark mockup cards MUST allow horizontal scroll on
narrow viewports rather than wrapping content lines. Line legibility MUST be preserved at all
breakpoints.

The top navigation MUST collapse to a full-screen overlay menu on viewports narrower than
768 px. A horizontal nav that overflows on mobile is prohibited.

**Rationale**: The primary audience includes mobile users sharing emotional moments in the
moment. Accessibility and responsive behaviour directly affect whether the platform delivers on
its emotional promise.

## Product Constraints

The following constraints are product decisions, not engineering preferences. They are
non-negotiable in v1 and MUST be reflected in every feature specification and implementation
plan.

- **No user accounts or login** — any feature requiring authentication is out of scope.
- **No deletion of submissions** — no user-facing or API-level delete mechanism.
- **Bilingual moderation** — profanity screening covers both English and Tagalog.
- **No financial data on platform** — donations are handled entirely off-platform via Ko-fi.
- **Image attachments** — maximum 5 MB per submission, accepted types JPEG, PNG, WebP, and GIF.
- **Character limits** — title: 3 to 30 characters (required); description: 0 to 200 characters
  (optional).
- **Database** — Supabase (Postgres + Storage) is the required persistence layer for v1.
- **Social sharing** — direct web-to-Story API posting is not available for Facebook or
  Instagram from a web application. Sharing MUST be implemented as client-side card image
  generation, file download, and native Web Share API with file (where supported).

## Design Governance

This constitution supersedes all other design and product practice documents for this project.
Amendments require: a written rationale, a version bump following semantic versioning rules
(MAJOR for removals or redefinitions, MINOR for additions, PATCH for clarifications), and
propagation to all dependent templates and specs.

All implementation plans MUST include a Constitution Check gate that verifies compliance with
the principles above before Phase 0 research proceeds, and re-verifies after Phase 1 design.
Violations that cannot be avoided MUST be documented in the plan's Complexity Tracking section
with an explicit justification and a simpler alternative that was considered and rejected.

Complexity decisions — introducing a pattern, dependency, or abstraction not called for by the
spec — MUST be justified in writing at the time of introduction. Unjustified complexity is a
constitution violation.

The project constitution is reviewed whenever a new feature spec is opened. If a new spec
introduces requirements that conflict with a principle, the principle MUST be amended before
the spec is accepted, not after implementation.

**Version**: 1.0.0 | **Ratified**: 2026-07-27 | **Last Amended**: 2026-07-27
