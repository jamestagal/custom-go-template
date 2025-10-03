**Goal:** Refactor component handling so that component ASTs are transformed and integrated *within* the main AST transformation process, eliminating the need for placeholder elements and regex-based rendering later.

---

**Checklist: Recursive Component Transformation**

**Phase 1: Setup & Prerequisites**

1.  **Component Registry Access:**
    *   [x] **Verify/Implement:** Ensure the `transformer` package has a reliable way to access the `componentTemplateRegistry` (currently populated in `cmd/server/main.go` via `transformer.RegisterComponent`).
    *   *Consider:* Maybe move registry management *into* the `transformer` package or pass it as an argument to `TransformAST`. For now, assume global access works, but flag for potential refactoring.
    *   *File(s):* `transformer/components.go`, potentially `transformer/transformer.go`, `cmd/server/main.go`

2.  **Remove Old Renderer Logic:**
    *   [x] **Delete/Comment Out:** Remove the entire `renderer.RenderComponents` function and its associated helpers (`getCompArgs`) from `renderer/component.go`.
    *   [x] **Update Callsites:** Remove any calls to `RenderComponents` from `renderer.Render` or elsewhere.
    *   *File(s):* `renderer/component.go`, `renderer/render.go`

**Phase 2: Core Component Transformation Logic (`transformer/components.go`)**

3.  **Refactor `transformComponent` Function Signature:**
    *   [x] Modify the function signature to accept the parent's data scope: `func transformComponent(node *ast.ComponentNode, parentDataScope map[string]any) []ast.Node`
    *   *File(s):* `transformer/components.go`

4.  **Implement `transformComponent` - Step 1: Lookup & Scope Init:**
    *   [x] Inside `transformComponent`, use `GetComponentTemplate(node.Name)` to retrieve the registered component's AST (`componentTemplate`).
    *   [x] Handle the case where the component is not found (log error, return placeholder/error node).
    *   [x] Create a *new, empty* `componentDataScope map[string]any` for this specific component instance.
    *   *File(s):* `transformer/components.go`

5.  **Implement `transformComponent` - Step 2: Process Component's Own Fence:**
    *   [x] Create a new helper function `collectComponentFenceData(fence *ast.FenceSection, scope map[string]any)`.
    *   [x] This helper should parse the component's *own* fence section (`componentTemplate.Template.FenceSection`).
    *   [x] Add declared variables and *default* prop values from the fence to `componentDataScope`. Use the `parseValue` helper (ensure it's accessible or copied).
    *   [x] Create another helper `extractFunctionsFromFence(rawContent string, scope map[string]any)` to find function definitions in the fence's raw content using regex.
    *   [x] Store extracted function definitions as *strings* in the `componentDataScope`.
    *   [x] Call `collectComponentFenceData` from `transformComponent`.
    *   *File(s):* `transformer/components.go`, potentially `transformer/scope.go` or `transformer/utils.go` for helpers.

6.  **Implement `transformComponent` - Step 3: Resolve Passed Props:**
    *   [x] Create a helper function `resolvePropValue(prop ast.ComponentProp, parentScope map[string]any) any`.
    *   [x] Inside `resolvePropValue`:
        *   If prop is dynamic (`prop.IsDynamic` or has `{}`), treat `prop.Value` as an expression name/path (e.g., `currentUser`, `product.price`). Retrieve the corresponding *value* from `parentScope`. If not found, log a warning and return `nil` or the expression string itself.
        *   If prop is shorthand (`prop.IsShorthand`), retrieve the value for `prop.Name` from `parentScope`.
        *   If prop is static, use `parseValue` on `prop.Value`.
    *   [x] Inside `transformComponent`, iterate through `node.Props` (the props passed *to* the component).
    *   [x] For each `passedProp`, call `resolvePropValue` using the `parentDataScope`.
    *   [x] Store the `resolvedValue` in the `componentDataScope` using `passedProp.Name` as the key (overwriting defaults from the fence if necessary).
    *   *File(s):* `transformer/components.go`

7.  **Implement `transformComponent` - Step 4: Transform Component Body:**
    *   [x] Get the component's body nodes (e.g., `componentTemplate.Template.RootNodes`, making sure to exclude the fence node itself using a helper like `filterOutFence`).
    *   [x] Recursively call `transformNodes` on these body nodes, passing the fully populated `componentDataScope` and `applyAlpineWrapper = false`. Store the result in `transformedChildren`.
    *   *File(s):* `transformer/components.go`

8.  **Implement `transformComponent` - Step 5: Add `x-data` Wrapper:**
    *   [x] Create a helper function `addComponentDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node`.
    *   [x] Inside `addComponentDataWrapper`:
        *   Call the *fixed* `alpineDataFormatter` (see Phase 3) on the `componentDataScope` to get the JS object string.
        *   Check if `nodes` represent a single root element. If yes, add the `x-data` attribute directly to that element.
        *   If `nodes` has multiple roots or isn't a single element, create a new wrapper `div` and add the `x-data` attribute to it, putting `nodes` as its children.
        *   Return the resulting single element (either the modified original or the new wrapper). Handle the case where `dataScope` might be empty (optional: skip `x-data`).
    *   [x] Call `addComponentDataWrapper` from `transformComponent` using `transformedChildren` and `componentDataScope`.
    *   *File(s):* `transformer/components.go`, `transformer/alpine.go`

9.  **Implement `transformComponent` - Step 6: Return Result:**
    *   [x] Return the final node(s) generated by `addComponentDataWrapper` from `transformComponent`.
    *   *File(s):* `transformer/components.go`

**Phase 3: Supporting Function Updates**

10. **Fix `alpineDataFormatter` and `formatGoValueToJS`:**
    *   [x] In `transformer/alpine.go`, replace the `json.Marshal`-based logic in `alpineDataFormatter` with direct string building using the `formatGoValueToJS` helper.
    *   [x] Ensure `formatGoValueToJS` correctly handles various Go types (string, bool, numbers, nil, slices, maps).
    *   [x] Crucially, ensure `formatGoValueToJS` identifies strings that are function definitions (using `isFunctionExpression`) and outputs them *without* surrounding quotes.
    *   *File(s):* `transformer/alpine.go`

11. **Update `transformNodes`:**
    *   [x] In `transformer/transformer.go`, modify the `case *ast.ComponentNode:` block within `transformNodes`.
    *   [x] Instead of creating a placeholder, it should now call the refactored `transformComponent(n, dataScope)`.
    *   [x] Append the *result* (which is `[]ast.Node`) of `transformComponent` to the `transformedNodes` list using `append(transformedNodes, componentNodes...)`.
    *   *File(s):* `transformer/transformer.go`

**Phase 4: Testing and Cleanup**

12. **Update Component Tests:**
    *   [x] Review `tests/alpine/components_test.go`, `tests/alpine/component_props_test.go`, `tests/alpine/dynamic_components_test.go`, `tests/components/component_test.go`.
    *   [x] The expected output should now reflect the *fully rendered component content* (with its own `x-data`) instead of just the `<div x-component="...">` placeholder. Adjust `expected` strings accordingly.
    *   [x] Ensure tests provide appropriate parent scopes (`props`) where needed for dynamic prop resolution.

13. **Run All Tests:**
    *   [ ] Execute the full test suite (`go test ./...`).
    *   [ ] Debug any remaining failures, likely related to scope interactions, `x-data` formatting, or prop resolution edge cases.

14. **Code Cleanup:**
    *   [ ] Remove any commented-out code related to the old rendering approach.
    *   [ ] Delete `.bak` files unless they contain specific logic still being evaluated.
    *   [ ] Ensure consistent logging levels and messages.

---

This checklist provides a detailed roadmap. Each step addresses a specific part of the required refactoring to achieve correct, AST-based recursive component transformation.