You said" The core issue is that Component:dynamic cannot resolve loop variables at transformation time. The solution is to use a workaround approach - pre-render the components instead of using Component:dynamic in a loop. Let me create a fix:" What is the point of implementing the dynamic component iteration if we can't use it as we designed and planned in the Spec? please explain what the actual limitation is and how and what approaches we would need to implement it have this feature work
You're absolutely right to call this out. Let me explain the actual limitation and what we need to implement to make this feature work as designed.
The Core Problem
The limitation is NOT that we can't use Component:dynamic in loops. The limitation is in how the transformation pipeline currently works: Current Flow:
Parser → AST → Transformer (resolves Component:dynamic) → Renderer
The Issue: When the transformer encounters <Component:dynamic name="{component.name}" /> inside a loop:
It tries to resolve {component.name} to a literal component name
But component.name is a loop variable that only exists at runtime
The transformer runs at compile/build time, not runtime
So it can't know what component.name will be
Why This Matters
This is actually a fundamental architectural question about when component resolution happens:
Option A: Compile-Time Resolution (Current)
Components are resolved during transformation
Results in static HTML with Alpine.js directives
Fast at runtime, but inflexible
Option B: Runtime Resolution (What We Need)
Components are resolved when the page loads
Requires Alpine.js to dynamically insert component HTML
Slower at runtime, but enables dynamic iteration
What We Actually Need to Implement
To make the spec work as designed, we need Runtime Component Resolution. Here are the approaches:
Approach 1: Alpine.js x-html with Component Registry (Recommended)
How it works:
Include ALL components in a JavaScript registry in the page
Use Alpine's x-html directive to inject component HTML at runtime
Alpine evaluates the loop variable and fetches the right component
Example transformation:
<!-- Template -->
{for component in components}
  <Component:dynamic name="{component.name}" {...component.fields} />
{/for}

<!-- Transforms to -->
<template x-for="component in components">
  <div x-html="$componentRegistry[component.name](component.fields)"></div>
</template>

<script>
  // Component registry with rendered HTML
  window.$componentRegistry = {
    'Hero2436': (props) => `<div class="hero">...</div>`,
    'FeatureCard': (props) => `<div class="card">...</div>`
  };
</script>
Pros:
Works with loop variables
Clean separation of concerns
Reusable across pages
Cons:
All components must be included in the page (larger payload)
Runtime HTML injection (slightly slower)
Approach 2: Server-Side Pre-rendering (Current Workaround)
How it works:
Server knows the content data at request time
Server iterates over components array and renders each one
Sends fully-rendered HTML to browser
Example (what we currently do in renderPlentiPage):
// Server reads content JSON
contentData := loadContent("/store-demo")
components := contentData["components"]

// Server iterates and renders each component
for _, comp := range components {
    componentHTML := renderComponent(comp.name, comp.fields)
    output += componentHTML
}
Pros:
Fast runtime (HTML already rendered)
Smaller payload (only used components)
SEO-friendly (content in initial HTML)
Cons:
Not truly "dynamic" (requires page reload to change)
Doesn't work for client-side component switching
Breaks the template → transform → render pipeline
Approach 3: Hybrid - Component Placeholder with Runtime Hydration
How it works:
Server pre-renders components with data attributes
Alpine.js hydrates them at runtime
Can swap components dynamically after initial load
Example:
<!-- Server renders -->
<template x-for="component in components">
  <div x-data="{compName: component.name, compProps: component.fields}">
    <div x-html="$componentRegistry[compName](compProps)"></div>
  </div>
</template>
What the Spec Actually Requires
Looking back at the spec in .agent-os/specs/2025-10-12-dynamic-component-iteration/spec.md, it requires:
✅ Parse Component:dynamic syntax
✅ Handle prop spreading from loop variables
❌ Runtime component resolution (NOT IMPLEMENTED)
❌ Component registry (NOT IMPLEMENTED)
Recommended Implementation Path
To make this feature work as designed, we need to implement:
Task 1: Component Registry System
Build a JavaScript registry of all components
Include rendered component HTML/templates
Make it accessible to Alpine.js
Task 2: Runtime Resolution Transformer
Transform Component:dynamic to x-html with registry lookup
Handle prop passing to component functions
Preserve Alpine reactivity
Task 3: Integration Testing
Test dynamic component switching
Test prop reactivity
Test nested dynamic components
My Mistake
I said "pre-render instead of using Component:dynamic" because I was looking at the immediate problem (making the store-demo page work) rather than the actual feature requirement (true dynamic component iteration). The workaround works but bypasses the template system entirely, which defeats the purpose of having Component:dynamic syntax in templates.
Question for You
Which approach do you want to pursue?
Full Runtime Resolution - Implement Approach 1 with component registry (true to the spec, but more complex)
Server-Side Only - Accept that dynamic iteration happens on the server (simpler, works now, but not truly dynamic)
Hybrid - Server pre-renders, but support runtime swapping for specific use cases
Let me know and I'll implement the proper solution!