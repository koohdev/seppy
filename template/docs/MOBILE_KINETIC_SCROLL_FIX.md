# Mobile Kinetic Scroll Fix: GSAP + SplitText + ScrollTrigger

This document explains a common performance pitfall where text reveal animations (using GSAP, SplitText, and ScrollTrigger) kill native mobile kinetic/inertial scrolling momentum, causing stutters, freezes, or repetitive re-trigger loops, and how to fix it in new projects.

---

## 1. The Problem

On mobile browsers (iOS Safari, Chrome, Android browsers), when a user swipes to scroll, the viewport height shifts dynamically as the address bar/navigation bar hides or shows. This viewport height shift triggers a window `resize` event.

If your text reveal system is set to re-split and recalculate text layout on _every_ resize event:

1. **Momentum Halted:** The DOM teardown (`revert()`) and layout reflow (`innerHTML` changes) force a layout flush on the browser main thread, instantly killing the momentum of the kinetic scroll.
2. **Infinite Flashing:** Text suddenly flashes back to hidden (because the animation resets to the start state, e.g., `yPercent: 110` or `opacity: 0.1`) and animates in again as the user scrolls.
3. **Double Handlers:** Re-triggering elements (via `toggleActions: "play none none reverse"`) constantly updates styling on scroll direction changes, compounding layout shifts.

---

## 2. The Solution Blueprint

To keep scrolling buttery smooth on mobile while maintaining text reveals, apply these three core architectural rules:

1. **Guard Resizes by Width only:** Viewport height changes (address bars) do not change the line wrapping of text. Only trigger text splitting and layout recalculation when `window.innerWidth` actually changes.
2. **Animate Once (`once: true`):** Entrance reveals look best when run only once. Set `once: true` by default. When the scroll triggers, the text animates in, stays visible, and the ScrollTrigger is killed/disabled.
3. **Track Revealed State with Attributes:** Use a `data-revealed="true"` attribute on elements that have completed their animation. When a legitimate width resize occurs (e.g., orientation changes), split the text to ensure responsive layout wrapping but skip re-running the animation entirely.

---

## 3. Implementation Code (Copy-Pasteable for New Projects)

Use this optimized helper in your new projects:

```javascript
import gsap from "gsap";
import { SplitText } from "./SplitText"; // or 'gsap/SplitText'
import { ScrollTrigger } from "gsap/ScrollTrigger";

gsap.registerPlugin(ScrollTrigger);

export function textReveal(scope = document, delay = 0) {
  if (typeof window === "undefined") return;

  const CONFIG = {
    lines: { duration: 1, stagger: 0.06, ease: "expo.out" },
    words: { duration: 1, stagger: 0.03, ease: "expo.out" },
    chars: { duration: 0.6, stagger: 0.01, ease: "expo.out" },
    scrollStart: "top 92%",
    scrubStart: "top 80%",
    scrubEnd: "top 20%",
    once: true, // Default to once to prevent repeat animations on scroll up/down
  };

  const allSplitEls = scope.querySelectorAll("[data-reveal]");

  let splitInstances = [];
  let animations = [];

  function initSplit() {
    // 1. Revert and clean up previous instances
    splitInstances.forEach((inst) => inst.revert());
    splitInstances = [];

    animations.forEach((anim) => {
      if (anim.scrollTrigger) anim.scrollTrigger.kill();
      anim.kill();
    });
    animations = [];

    const isMobile = window.matchMedia("(max-width: 991px)").matches;

    allSplitEls.forEach((el) => {
      const htmlEl = el;
      let splitType = htmlEl.getAttribute("data-reveal") || "lines";

      // Downgrade heavy splitting on mobile to words/lines for performance
      if (isMobile && splitType === "chars") {
        splitType = "words";
      }

      const c = CONFIG[splitType];
      if (!c) return;

      // Define SplitText configurations
      let type = "lines";
      let mask = "lines";
      let linesClass = "line";
      let wordsClass = "";
      let charsClass = "";

      if (splitType === "words") {
        type = "words, lines";
        mask = "words";
        wordsClass = "word";
      } else if (splitType === "chars") {
        type = "chars, words, lines";
        mask = "chars";
        charsClass = "char";
        wordsClass = "word";
      }

      // Initialize SplitText
      const inst = SplitText.create(htmlEl, {
        type,
        mask,
        autoSplit: true,
        ...(linesClass && { linesClass }),
        ...(wordsClass && { wordsClass }),
        ...(charsClass && { charsClass }),
        onSplit(instance) {
          // Rule 3: Skip animations for already-revealed elements (preserves final state on resize)
          if (htmlEl.getAttribute("data-revealed") === "true") {
            return;
          }

          const duration = parseFloat(htmlEl.dataset.duration) || c.duration;
          const stagger = parseFloat(htmlEl.dataset.stagger) || c.stagger;
          const elDelay = parseFloat(htmlEl.dataset.delay) || 0;
          const ease = htmlEl.dataset.ease || c.ease;
          const scrollMode = htmlEl.getAttribute("data-scroll");
          const useScroll = htmlEl.hasAttribute("data-scroll");
          const useScrub = scrollMode === "scrub";
          const targets = instance[splitType];

          const tweenVars = {
            yPercent: 110,
            duration,
            stagger,
            delay: useScroll ? elDelay : elDelay + delay,
            immediateRender: true,
            ease,
            onComplete: () => {
              // Mark element as revealed once animation completes
              htmlEl.setAttribute("data-revealed", "true");
            },
          };

          if (useScrub) {
            if (isMobile) {
              // Mobile: use triggerActions instead of scrub to eliminate per-frame math
              tweenVars.scrollTrigger = {
                trigger: htmlEl,
                start: CONFIG.scrubStart,
                toggleActions: "play none none none",
              };
            } else {
              tweenVars.scrollTrigger = {
                trigger: htmlEl,
                start: CONFIG.scrubStart,
                end: CONFIG.scrubEnd,
                scrub: true,
                onLeave: (self) => {
                  self.kill(false); // Disable ScrollTrigger but keep revealed state
                  htmlEl.setAttribute("data-revealed", "true");
                },
              };
            }
          } else if (useScroll) {
            const start =
              scrollMode && scrollMode !== "true"
                ? scrollMode
                : CONFIG.scrollStart;
            tweenVars.scrollTrigger = {
              trigger: htmlEl,
              start: `clamp(${start})`,
              once: true, // Fire only once
            };
          }

          const anim = gsap.from(targets, tweenVars);
          animations.push(anim);
        },
      });
      splitInstances.push(inst);
    });
  }

  // First run
  initSplit();

  // Rule 1: Guard resizes by window width to protect mobile scrolling
  let lastWidth = window.innerWidth;
  let resizeTimeout;

  const handleResize = () => {
    const currentWidth = window.innerWidth;
    if (currentWidth === lastWidth) return; // Ignore height shifts (address bar changes)
    lastWidth = currentWidth;

    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(() => {
      initSplit();
      ScrollTrigger.refresh();
    }, 250);
  };

  window.addEventListener("resize", handleResize);

  // Return cleanup function to call on component unmount
  return () => {
    window.removeEventListener("resize", handleResize);
    splitInstances.forEach((inst) => inst.revert());
    animations.forEach((anim) => {
      if (anim.scrollTrigger) anim.scrollTrigger.kill();
      anim.kill();
    });
  };
}
```

---

## 4. Why This Works:

1. **0% DOM mutations during normal mobile scrolling:** Prevents frame drops, scroll freezes, or loss of kinetic scroll momentum.
2. **Reduces layout updates:** The `data-revealed` guard guarantees that once an element reveals, it stays static—even on rotation resizes.
3. **No reverse-animation layout overhead:** Eliminates scroll direction checking calculations.
