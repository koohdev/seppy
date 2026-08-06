Here is the updated documentation with the link included.

# [NumberFlow](https://number-flow.barvian.me/)

An animated number component. Dependency-free. Accessible. Customizable. Created by **Maxwell Barvian**.

---

## Basic Usage

`<NumberFlow>` will automatically transition when the `value` prop changes.

```tsx
import NumberFlow from "@number-flow/react";

function Example() {
  return <NumberFlow value={123} />;
}
```

> 💡 **Learn to build components like NumberFlow:**
> _"Emil Kowalski's Animations on the Web course taught me everything I know about UI animation. I can't imagine having built NumberFlow without it."_
> — **Maxwell Barvian**, NumberFlow creator

---

## Props

### `format`

- **Type:** `Intl.NumberFormatOptions`
- **Description:** Formatting options for the number.

```tsx
<NumberFlow format={{ notation: "compact" }} value={value} />
```

### `locales`

- **Type:** `Intl.LocalesArgument`
- **Description:** The locale(s) for the number.

### `prefix` & `suffix`

- **Type:** `string`
- **Description:** A custom prefix or suffix for the number.

```tsx
<NumberFlow
  value={value}
  format={{
    style: "currency",
    currency: "USD",
    trailingZeroDisplay: "stripIfInteger",
  }}
  suffix="/mo"
/>
```

### Timings

There are three props to customize the animation timings. Each accepts an `EffectTiming` object:

```tsx
<NumberFlow
  // Used for layout-related transforms:
  transformTiming={{ duration: 750, easing: "linear(...)" }}
  // Used for the digit spin animations. (Falls back to transformTiming if unset):
  spinTiming={{ duration: 750, easing: "linear(...)" }}
  // Used for fading in/out characters:
  opacityTiming={{ duration: 350, easing: "ease-out" }}
/>
```

_Tip: For spring-based easings, use tools like Kevin Grajeda’s generator or easing.dev._

### `trend`

- **Type:** `number | ((oldValue: number, value: number) => number)`
- **Default:** `(oldValue, value) => Math.sign(value - oldValue)`
- **Description:** Controls the direction of the digits:
- `+1`: Digits always go up.
- `0`: Each digit goes up if it increases and down if it decreases (useful for shifting numbers without conveying a trend).
- `-1`: Digits always go down.

### `isolate`

- **Type:** `boolean`
- **Default:** `false`
- **Description:** Isolates `<NumberFlow>` transitions from other layout changes in the same update. Has no effect when nested inside a `<NumberFlowGroup>`.

### `animated`

- **Type:** `boolean`
- **Default:** `true`
- **Description:** Set to `false` to disable animations and instantly complete any running ones.

### `digits`

- **Type:** `Record<number, { max?: number }>`
- **Description:** Configure digits based on their position in the number (e.g., for `342.5`, positions are `3` [2], `4` [1], `2` [0], `.` [-1]). Helpful for countdown displays (e.g., ensuring `59 -> 00`).
- _Note: `digits` is not reactive to save on bundle size._

### `respectMotionPreference`

- **Type:** `boolean`
- **Default:** `true`
- **Description:** Set to `false` to animate regardless of the user’s OS-level reduced motion preferences.

### `plugins`

- **Type:** `Plugin[]`
- **Description:** Plugins to apply to the component. Currently, the only plugin is `continuous`, which makes number transitions appear to scroll continuously through in-between numbers. _(Has no effect if `trend` is `0`.)_

### `willChange`

- **Type:** `boolean`
- **Default:** `false`
- **Description:** Applies CSS `will-change` properties to relevant elements. Recommended if your number updates extremely frequently or you experience micro-stutter/repositioning on complete.
- _Warning: Excessive use of `will-change` can consume high memory._

### `nonce`

- **Type:** `string`
- **Description:** Passes a CSP nonce through to NumberFlow’s inline `<style>` elements during SSR and hydration.

```tsx
<NumberFlow value={123} nonce={myNonce} />
```

### `onAnimationsStart`

- **Type:** `(e: CustomEvent) => void`
- **Description:** Triggered when update animations begin.

### `onAnimationsFinish`

- **Type:** `(e: CustomEvent) => void`
- **Description:** Triggered when update animations complete.

---

## Styling

`<NumberFlow>` uses a **custom element** under the hood and exposes standard CSS parts for styling:

```css
/* Example targets */
::part(digit) {
  /* custom styling */
}
```

> ⚠️ **Note:** Changing the `font-size` of digits dynamically is difficult due to the underlying CSS techniques. Additionally, `::part` styles may cause a flash of unstyled content (FOUC) in old browsers.

### CSS Variables

| Property                    | Default        | Description                                                                           |
| --------------------------- | -------------- | ------------------------------------------------------------------------------------- |
| `--number-flow-mask-height` | `.25em`        | Adjusts the height of the gradient fade-out mask at the top/bottom edges.             |
| `--number-flow-mask-width`  | `.5em`         | Adjusts the width of the gradient fade-out mask at the sides.                         |
| `line-height`               | `1`            | Adjusts vertical spacing between digits during spin animations (0.85 is recommended). |
| `font-variant-numeric`      | `tabular-nums` | Ensures all numbers are the same width to prevent layout shifts.                      |

---

## Grouping

If multiple independent `<NumberFlow>` elements affect each other's positions layout-wise, wrap them in a `<NumberFlowGroup>` to perfectly sync their layout transitions:

```tsx
<NumberFlowGroup>
  <NumberFlow value={124.23} />
  <NumberFlow value={5.64} prefix="+" suffix="%" />
</NumberFlowGroup>
```

_Note: `<NumberFlowGroup>` does not render a DOM element or accept props._

---

## Hooks

### `useCanAnimate`

- **Signature:** `useCanAnimate(opts?: { respectMotionPreference?: boolean }): boolean`
- **Description:** Returns `true` if the browser supports required animation features and the user hasn't requested reduced motion.

---

## Limitations

- **Scientific and engineering notations** are not supported.
- **Non-Latin digits** and **RTL (right-to-left) locales** are not currently supported.
- **Backgrounds & borders** on `<NumberFlow>` will not scale smoothly during transitions. For smooth container scaling, use a dedicated layout animation library like **Motion for React**.

---

_Built by [Max Barvian](https://number-flow.barvian.me/). Heavily inspired by the **Family** app. Mask-image technique by **Artur Bień**. Digit looping technique by **Sam Selikoff**._
