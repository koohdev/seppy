# Text Slide-Up Hover Animation (Rolling Reveal)

This effect is widely known as a **Text Slide-Up Reveal**, **Rolling Text Hover Transition**, or **Duplicate Slide-Up Animation**.

It works by masking a container with `overflow-hidden` and translating two vertically stacked, identical sets of content (text and icons) upward when the parent button is hovered.

---

## Tailwind CSS Implementation

Here is a highly optimized, clean Tailwind implementation using standard utilities:

```tsx
<button className="group relative overflow-hidden rounded-full bg-blue-600 px-6 py-3 text-white">
  {/* The container needs a fixed height or a line-height restriction to mask perfectly */}
  <span className="relative block h-[1.5em] overflow-hidden">
    {/* Default State (Slides UP on hover) */}
    <span className="flex items-center gap-2 transition-transform duration-300 ease-out group-hover:-translate-y-full">
      <span>Get Started</span>
      <svg
        className="w-4 h-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M14 5l7 7m0 0l-7 7m7-7H3"
        />
      </svg>
    </span>

    {/* Hover State (Starts below, slides UP into position on hover) */}
    <span className="absolute left-0 top-0 flex items-center gap-2 translate-y-full transition-transform duration-300 ease-out group-hover:translate-y-0">
      <span>Get Started</span>
      <svg
        className="w-4 h-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M14 5l7 7m0 0l-7 7m7-7H3"
        />
      </svg>
    </span>
  </span>
</button>
```

---

## Core Technical Requirements for the AI

When implementing this, ensure you apply the following structural rules:

1. **Parent Wrapper:** Must have `group` and `relative` to coordinate the hover state and absolute positioning.
2. **Masking Container:** Must have `overflow-hidden` and a height matching the line-height (e.g., `h-[1.5em]`) to hide the duplicate text.
3. **Primary Content:** Must use `transition-transform duration-[speed] group-hover:-translate-y-full`.
4. **Secondary (Duplicate) Content:** Must use `absolute top-0 left-0 translate-y-full transition-transform duration-[speed] group-hover:translate-y-0`.
