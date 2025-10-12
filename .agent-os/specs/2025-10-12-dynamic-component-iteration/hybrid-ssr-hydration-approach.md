 Here’s the straight answer, then the best path forward.

# What’s actually limiting us

* **Where resolution happens.** Our current pipeline resolves `<Component:dynamic …>` **during the transform step (compile time)**.
* **Loop scope lives at runtime.** In a `{for component in components}` loop, `component.name` is only known **at runtime** in the browser (or at request-time on the server), not at compile time.
* **Result:** When the transformer tries to turn
  `<Component:dynamic name="{component.name}" …/>`
  into a concrete component, it can’t — the value doesn’t exist yet. That’s why it “fails” inside loops (and any other runtime-only scope).

So the bug isn’t “you can’t use `Component:dynamic` in loops”; it’s “you can’t try to **resolve** it at **compile time** if its name comes from a runtime value.” The feature as specced implicitly requires **runtime resolution**.

# Why we built what we built (and why it felt wrong)

* We initially optimized for **compile-time/static** transformation (fast render, smaller payloads).
* That works for `name="Header"` but breaks for `name={component.name}` in loops.
* The server-side workaround did render pages, but it **bypassed the new template feature**, which defeats the purpose of having `<Component:dynamic>` in the first place. Your pushback is correct.

# Viable approaches (trade-offs)

1. **Pure runtime resolution (client-side)**

   * **How:** Emit a placeholder and let Alpine resolve and inject the right component HTML from a **component registry** using `x-html`.
   * **Pros:** Matches the spec, supports loops & switching after load.
   * **Cons:** Larger JS; HTML comes via client injection; slightly slower TTI; must ship registry.

2. **Pure server-side resolution (request-time)**

   * **How:** Resolve components on the server using the content JSON and output final HTML.
   * **Pros:** SEO perfect, fast first paint, small JS.
   * **Cons:** Not truly dynamic after load; changing component selection requires a reload; undermines the point of `<Component:dynamic>` in templates (you’re doing it in Go instead).

3. **Hybrid (recommended)** — **SSR first**, then **runtime-capable hydration**

   * **How:**

     * At request time, the server renders the initial component list (SEO + fast paint).
     * It embeds small **descriptors** (`data-comp`, `data-props`) next to each rendered block.
     * Alpine + registry hydrate those blocks so future **runtime swaps** (loop updates, filters, “change layout” UI) re-render the correct component client-side without reload.
   * **Pros:** Best of both worlds — SEO + performance + true runtime iteration/switching.
   * **Cons:** Slightly more implementation work (need both initial SSR and client hydrator).

# What changes we need (concretely)

Regardless of the variant, we must stop trying to **fully resolve** dynamic components in the transformer when the name is an expression that depends on runtime scope. Instead:

## A. Adjust the transformer’s behavior

* **If `name=` is a static string** → we may keep compile-time resolution (safe optimization).
* **If `name=` is an expression** (e.g., `component.name`) → **do not resolve** at transform time. Emit a **runtime-capable node**.

### Output pattern (runtime-capable node)

Inside a loop, transform to something like:

```html
<template x-for="component in components" :key="component.id || component.name">
  <div
    class="dyn-comp"
    x-data="{ compName: component.name, compProps: component.fields }"
    x-init="$renderDynamic($el, compName, compProps)">
    <!-- SSR fallback / placeholder will live here (see Hybrid) -->
  </div>
</template>
```

…and provide a tiny runtime:

```html
<script>
  window.$componentRegistry = {
    Hero2436: (props) => `<section id="hero-2436">...</section>`,
    FeatureCard: (props) => `<div class="card">...</div>`,
    // …
  };

  // Minimal re-render helper (idempotent)
  window.$renderDynamic = (el, name, props) => {
    const fn = window.$componentRegistry?.[name];
    if (!fn) {
      el.innerHTML = `<!-- Warning: Component '${name}' not found -->`;
      return;
    }
    el.innerHTML = fn(props || {});
  };
</script>
```

## B. Implement **prop merging** with spreads (as specced)

* Evaluate spreads **at the point of runtime render**, with **rightmost-wins** precedence:

  1. left→right spreads
  2. then left→right regular props
  3. later items override earlier ones
* In the transformer, for dynamic cases, serialize the merge **plan** alongside the node (so the hydrator can execute the same precedence on the client). For SSR/hybrid, do the same merging on the server too (match behavior).

## C. Add a **component registry**

* **Build step (Go):** collect all component templates into a registry map keyed by name.
* **Ship strategy:**

  * **Hybrid optimization:** only embed entries for components actually used on the page + any whitelisted globals (tree-shake rest).
  * Allow code-split bundles later if we add lazy loading.

## D. (Hybrid) Do **SSR of the initial state** + client hydration

* On the server, when we already know `components` at request time:

  * Render the correct component HTML into each placeholder **and** attach:

    * `data-comp="Hero2436"`
    * `data-props='{…json…}'`
  * On the client `x-init`, compare/replace using `$renderDynamic` only if needed (or simply accept it as already correct).
* This yields:

  * Fast, SEO-friendly initial page.
  * Real runtime switching later if the `components` array changes.

### Example (hybrid final HTML)

```html
<template x-for="component in components" :key="component.id || component.name">
  <div
    class="dyn-comp"
    x-data="{ compName: component.name, compProps: component.fields }"
    x-init="$renderDynamic($el, compName, compProps)"
    data-comp="Hero2436"
    data-props='{"title":"Main Title","buttonLink":"/contact"}'>
    <!-- SSR-rendered HTML is here already -->
    <section id="hero-2436">…</section>
  </div>
</template>
```

# Addressing your “what’s the point?” concern

With the **hybrid** solution:

* `<Component:dynamic>` **stays in templates** (no server bypass),
* It **works in loops** and supports **spread props**,
* It renders **instantly** with SSR, and
* It can **change dynamically** after load (filters, AB tests, UI toggles, personalization).

That **is** the point of implementing the feature in the template system: authors get a declarative, Svelte-like idiom that’s both SEO-friendly and dynamic.

# What to change in our spec & tasks (practical plan)

* **Transformer** (update):

  * Detect runtime expressions in `name=` and **emit runtime-capable nodes** instead of failing to resolve.
  * Carry a compiled “merge recipe” for spreads/props so SSR and client share semantics.

* **Renderer** (extend):

  * **SSR path**: if `components` data is available, render the initial HTML + `data-*` descriptors.
  * Otherwise output empty wrapper with `data-comp`/`data-props` for client fill.

* **Client runtime** (new tiny module):

  * `$componentRegistry` (pluggable)
  * `$renderDynamic(el, name, props)`
  * Optional: a watcher to re-render if `compName`/`compProps` change.

* **Performance safeguards**:

  * Tree-shake registry to “used + allowlist”.
  * Optionally gate very large components behind `IntersectionObserver` (lazy render) in a future patch.
  * Keep Alpine responsibilities tiny (no heavy VDOM).

* **Error handling** (as specced):

  * Unknown component → HTML comment warning.
  * Spread not an object → console error + skip.
  * Cycles → detect in SSR traversal with a depth cap.

* **Testing** (map to your unchecked items):

  * E2E for mixed SSR + runtime swap (6.1–6.6).
  * Docs to clarify the two modes (7.x), especially that **static `name`** can still be compile-time resolved.

# Bottom line (recommendation)

Go with the **Hybrid** approach now:

* Keeps the promise of the **template feature** (authors use `<Component:dynamic>` in loops as designed).
* Preserves **SEO & performance** with SSR.
* Enables **true runtime iteration** and switching later — no reloads.
* Contains payload growth via **registry tree-shaking**.

If you’re happy, I’ll align the remaining Tasks 6–7 to this hybrid design (no more compile-time resolution attempts for runtime names) and wire the tiny runtime + SSR descriptors so your current pages get both speed and dynamism.
