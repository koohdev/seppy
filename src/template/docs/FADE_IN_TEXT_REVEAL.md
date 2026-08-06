# Text Reveal 02

## 1. What this resource is

This resource is a GSAP-based text reveal system. It opts text elements into the effect with `data-reveal-02`, splits them into lines, words, or characters with SplitText, and fades each piece in from `opacity: 0.1`.

default: lines and scrub

It supports:

- load-time reveals
- scroll-triggered reveals
- scrubbed scroll reveals
- manual split-only mode for custom timelines
- re-running the helper for dynamically injected content

Before writing any code, inspect the target project and confirm:

- where the text markup lives
- whether client-side JavaScript can run after render
- whether GSAP, SplitText, and ScrollTrigger can be loaded in that stack
- whether the project uses client-side page swaps or injected content that will need re-initialization

If the target project cannot support this kind of client-side DOM mutation and JS animation, stop and explain the incompatibility instead of forcing a broken port.

## 2. Compatible targets

- Vanilla HTML/CSS/JS sites
- Astro, React, Vue, Svelte, Next, Nuxt, and similar apps that can run client-side JS after the text is rendered
- CMS or templated sites where you can keep the required `data-*` attributes and run a helper on the client

## 3. Not compatible with this exact implementation

- Projects with no client-side JavaScript
- Environments where GSAP and its plugins cannot be loaded
- Contexts that strip the required `data-*` attributes
- Contexts that cannot safely let SplitText inject wrapper elements around the text
- SSR-only rendering paths where the text must never be mutated after render

If the project falls into one of those cases, do not fake a partial match. Explain what is incompatible and what would need to change.

## 4. Source stack and dependencies

- The demo shell uses Astro, but the effect itself is plain client-side JavaScript.
- Core dependency: `gsap`
- Core GSAP plugins: `SplitText` and `ScrollTrigger`
- `gsap` drives the animation
- `SplitText` creates the `.line`, `.word`, and `.char` elements (no mask wrappers)
- `ScrollTrigger` powers `data-scroll` and `data-scroll="scrub"`

Do not treat the Astro layout, demo sections, background colors, or preview/reset scripts as part of the core resource. The resource is the reveal system itself.

## 5. Preserve or map these selectors and attributes

Preserve these exact names unless the target project absolutely requires a rename. If you rename anything, rename it consistently across markup, CSS, and JS.

- `[data-reveal-02]` = required opt-in attribute on the text element itself
- `data-reveal-02="lines"` = split into lines
- `data-reveal-02="words"` = split into words and lines
- `data-reveal-02="chars"` = split into characters, words, and lines
- `data-scroll` = threshold-based scroll trigger using the helper defaults
- `data-scroll="scrub"` = scrubbed scroll-driven reveal
- `data-duration`, `data-stagger`, `data-delay`, `data-ease` = per-element overrides
- `data-once` = per-element override for scroll replay behavior
- `data-manual` = split the text but skip the automatic animation
- `.line`, `.word`, `.char` = generated SplitText elements used as animation targets (no mask wrappers in this resource)
- `textReveal02(scope = document, delay = 0, { ignoreManual = false } = {})` = helper signature to preserve or map carefully

## 6. Required markup usage

Put the attribute on the actual text element, not on a parent wrapper.

```html
<p data-reveal-02="lines">Your text goes here.</p>

<p data-reveal-02="words" data-scroll>
  This one animates when it enters the viewport.
</p>

<p data-reveal-02="chars" data-scroll="scrub">
  This one follows the scroll position.
</p>
```

Important markup rules:

- The effect attaches to the text element itself.
- The surrounding demo sections from the source are only showcase structure, not required.
- Only use the real supported values: `lines`, `words`, `chars`, or the real scroll values shown above.

## 7. Required CSS and hidden-state setup

This resource does not use mask wrappers. Only the hidden-state rule is required.

```css
[data-reveal-02] {
  visibility: hidden;
}
```

What this CSS does:

- hides the text before JS runs so it does not flash in its final state

## 8. Required JS setup and init timing

Preserve this helper logic closely. The split configuration, per-element overrides, and ScrollTrigger wiring are part of the resource contract. This resource does not use SplitText's mask option; it animates opacity on the generated `.line`, `.word`, or `.char` elements directly.

```js
import gsap from "gsap";
import { SplitText } from "gsap/SplitText";
import { ScrollTrigger } from "gsap/ScrollTrigger";

gsap.registerPlugin(SplitText, ScrollTrigger);

export function textReveal02(
  scope = document,
  delay = 0,
  { ignoreManual = false } = {},
) {
  const CONFIG = {
    lines: { duration: 0.04, stagger: 0.03, ease: "power1.out" },
    words: { duration: 0.04, stagger: 0.03, ease: "power1.out" },
    chars: { duration: 0.04, stagger: 0.03, ease: "power1.out" },
    scrollStart: "top 85%",
    scrubStart: "top 80%",
    scrubEnd: "top 20%",
    once: true,
    markers: false,
  };

  const allSplitEls = scope.querySelectorAll("[data-reveal-02]");
  const autoEls = ignoreManual
    ? [...allSplitEls]
    : [...allSplitEls].filter((el) => !el.hasAttribute("data-manual"));

  gsap.set(autoEls, { visibility: "visible" });

  allSplitEls.forEach((el) => {
    const splitType = el.getAttribute("data-reveal-02");
    const c = CONFIG[splitType];
    if (!c) return;

    let type, linesClass, wordsClass, charsClass;
    switch (splitType) {
      case "lines":
        type = "lines";
        linesClass = "line";
        break;
      case "words":
        type = "words, lines";
        wordsClass = "word";
        linesClass = "line";
        break;
      case "chars":
        type = "chars, words, lines";
        charsClass = "char";
        wordsClass = "word";
        linesClass = "line";
        break;
      default:
        return;
    }

    if (!ignoreManual && el.hasAttribute("data-manual")) {
      SplitText.create(el, {
        type,
        autoSplit: true,
        ...(linesClass && { linesClass }),
        ...(wordsClass && { wordsClass }),
        ...(charsClass && { charsClass }),
      });
      return;
    }

    const scrollMode = el.getAttribute("data-scroll");
    const useScroll = el.hasAttribute("data-scroll");
    const useScrub = scrollMode === "scrub";

    SplitText.create(el, {
      type,
      autoSplit: true,
      ...(linesClass && { linesClass }),
      ...(wordsClass && { wordsClass }),
      ...(charsClass && { charsClass }),
      onSplit(instance) {
        const durationValue = parseFloat(el.dataset.duration);
        const staggerValue = parseFloat(el.dataset.stagger);
        const delayValue = parseFloat(el.dataset.delay);
        const duration = Number.isNaN(durationValue)
          ? c.duration
          : durationValue;
        const stagger = Number.isNaN(staggerValue) ? c.stagger : staggerValue;
        const elDelay = Number.isNaN(delayValue) ? 0 : delayValue;
        const ease = el.dataset.ease || c.ease;

        const targets = instance[splitType];
        const once = el.hasAttribute("data-once")
          ? el.getAttribute("data-once") !== "false"
          : CONFIG.once;

        const tween = {
          opacity: 0.1,
          duration,
          stagger,
          delay: useScroll ? elDelay : elDelay + delay,
          immediateRender: true,
          ease,
        };

        if (useScrub) {
          tween.scrollTrigger = {
            trigger: el,
            start: CONFIG.scrubStart,
            end: CONFIG.scrubEnd,
            scrub: true,
            markers: CONFIG.markers,
            ...(once && { onLeave: (self) => self.kill(false) }),
          };
        } else if (useScroll) {
          const start = scrollMode || CONFIG.scrollStart;
          tween.scrollTrigger = {
            trigger: el,
            start: `clamp(${start})`,
            markers: CONFIG.markers,
            ...(once
              ? { once: true }
              : { toggleActions: "play none none reverse" }),
          };
        }

        return gsap.from(targets, tween);
      },
    });
  });
}
```

Keep these setup behaviors:

- Register both `SplitText` and `ScrollTrigger`.
- Run the helper after the target text is rendered on the client.
- Keep `autoSplit: true`.
- Keep the `onSplit()` callback and return the GSAP tween from it.
- Preserve the `clamp(${start})` behavior for threshold-based scroll starts.
- Preserve the split target lookup via `instance[splitType]`.
- Do not add a `mask` option to SplitText — this resource animates opacity directly on the generated elements.

For first load, the source uses `document.fonts.ready` before the first call:

```js
document.addEventListener("DOMContentLoaded", () => {
  document.fonts.ready.then(() => {
    textReveal02();
  });
});
```

If custom fonts do not visibly change line breaks, it is also valid to call the helper directly on DOMContentLoaded:

```js
document.addEventListener("DOMContentLoaded", () => {
  textReveal02();
});
```

If the target stack does not use ESM imports, adapt the loading style, but keep the helper logic the same.

## 9. Customization options

The main customization points are the shared `CONFIG` object and the per-element data attributes.

- `lines`, `words`, `chars` each have their own default `duration`, `stagger`, and `ease`
- `scrollStart` controls the default threshold-based scroll trigger start
- `scrubStart` and `scrubEnd` control the scrub range
- `once` defaults scroll-triggered reveals to play once
- `markers` is a debug flag and should usually stay `false`

Per-element overrides:

- `data-duration="0.8"` overrides the item duration
- `data-stagger="0.05"` overrides the stagger
- `data-delay="0.2"` overrides the delay for that element
- `data-ease="power2.out"` overrides the GSAP ease
- `data-once="false"` changes scroll replay behavior for that element

Behavior details to preserve:

- In non-scroll mode, the helper-level `delay` is added to the element's own `data-delay`
- In scroll mode, the helper-level `delay` is ignored and only the element's own `data-delay` is used
- For threshold-based scroll reveals, `once: true` uses ScrollTrigger's `once` behavior
- For threshold-based scroll reveals, `data-once="false"` switches to `toggleActions: "play none none reverse"`
- For scrub mode, `once: true` kills the ScrollTrigger on leave so the final state stays locked

## 10. Dynamic-content and page-transition re-init pattern

This helper is designed to be run again for new content.

Preserve these meanings:

- `scope` = container to query within instead of always searching the whole document
- `delay` = extra delay in seconds for auto-playing elements
- `ignoreManual` = animate normally manual elements too

Typical re-init examples:

```js
textReveal02(document);

const nextContainer = document.querySelector("#new-content");
textReveal02(nextContainer, 0.3);

textReveal02(document, 0, { ignoreManual: true });
```

Use that same pattern after router-driven page swaps, partial page updates, or dynamically injected sections. Do not rerun the helper blindly against the whole document if the target stack gives you the new container directly.

## 11. Important notes and gotchas

- **React Boolean Attribute Gotcha:** In React/Next.js, writing a custom attribute like `data-scroll` without a value compiles into the HTML DOM as `data-scroll="true"`. Therefore, `htmlEl.getAttribute('data-scroll')` returns the string `"true"`. If you evaluate `scrollMode || CONFIG.scrollStart`, it evaluates to `"true"`, which is an invalid ScrollTrigger coordinate. GSAP falls back to its default `"top top"` (scroller-start at `y=0`), making it trigger late or not at all. Always sanitize: `const start = (scrollMode && scrollMode !== 'true' && scrollMode !== '') ? scrollMode : CONFIG.scrollStart;`.
- **GSAP `once: true` with `fromTo` Tweens:** Setting `once: true` inside a ScrollTrigger config can cause GSAP to immediately destroy the trigger upon enter. For `fromTo` tweens, this can freeze the animation progress at `0`, keeping the element locked at its initial state (`opacity: 0.1`). Use `toggleActions: once ? 'play none none none' : 'play none none reverse'` instead to safely animate once without killing the ScrollTrigger object mid-animation.
- **SplitText on Parent Wrappers:** Never attach `data-reveal-02` to parent layout wrappers (e.g. `div` containers wrapping grid lists). `SplitText` reads the `textContent` of the target (concatenating all child columns) and clears the innerHTML to render line wraps, which destroys the list/grid tags completely. Always attach reveal attributes directly to the actual text tags (like `li`, `p`, or `h2`), not their parent container.
- The hidden-state CSS is required or the text will flash before JS reveals it.
- `data-manual` still runs SplitText, but it does not animate or reveal the text automatically.
- The generated `.line`, `.word`, and `.char` elements are part of the effect and can be used in manual GSAP timelines. For manual mode, set initial state with `gsap.set(els, { opacity: 0.1 })` and animate to `opacity: 1`.
- If line-based text reflows after fonts load, initialize after `document.fonts.ready` or otherwise account for that in the target stack.
- SplitText's default accessibility behavior is `aria: "auto"`. Preserve that default unless the target project already has its own accessibility strategy for split text.

## 12. What not to confuse with the core reveal

Do not port these as if they are part of the effect:

- demo page section layouts
- showcase copy
- background color swaps
- Astro-specific layout structure
- preview/reset helper scripts
- unrelated page-transition code

Implement the reveal system first. The demo shell is not the deliverable.

## 13. Deliverable

If the target project is compatible, produce:

- the final integrated code in the correct files for that stack
- any selector or template mapping needed for that codebase
- the required CSS and helper logic, adapted but faithful to the original behavior
- any brief integration notes that are genuinely necessary for the target project

If the target project is not compatible, stop and explain why instead of forcing a broken approximation.
