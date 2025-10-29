/**
 * Runtime Component Resolution for Alpine.js
 *
 * This module provides the client-side runtime for dynamic component rendering.
 * It loads a component registry and renders components at runtime based on
 * component names that are only known during Alpine.js execution (e.g., from loops).
 *
 * Pattern: Client-Side Runtime Module [Cognitive Load: 22]
 * - Registry loading: 10
 * - Magic function: 12
 *
 * Usage:
 *   <div x-data="{compName: 'Hero2436', compProps: {...}}"
 *        x-init="$renderDynamicComponent($el, compName, compProps)">
 *   </div>
 */

// Registry cache - loaded once and reused
let componentRegistry = null;
let registryLoadPromise = null;

// Configuration
const REGISTRY_URL = '/js/component-registry.js';
const MAX_RETRY_ATTEMPTS = 3;
const INITIAL_RETRY_DELAY = 500; // milliseconds

/**
 * Load component registry from server
 *
 * Pattern: Async Loading with Retry [Cognitive Load: 10]
 * - Dynamic import: 3
 * - Retry logic: 5
 * - Error handling: 2
 *
 * Implements exponential backoff retry strategy:
 * - Attempt 1: immediate
 * - Attempt 2: 500ms delay
 * - Attempt 3: 1000ms delay
 *
 * @returns {Promise<Object>} Component registry object
 * @throws {Error} If all retry attempts fail
 */
async function loadComponentRegistry() {
    // If already loaded, return cached registry
    if (componentRegistry) {
        console.log('Component registry already loaded (cached)');
        return componentRegistry;
    }

    // If loading is in progress, wait for it
    if (registryLoadPromise) {
        console.log('Component registry loading in progress, waiting...');
        return registryLoadPromise;
    }

    // Start loading
    console.log('Loading component registry...');

    registryLoadPromise = (async () => {
        let lastError = null;

        // Retry loop with exponential backoff
        for (let attempt = 1; attempt <= MAX_RETRY_ATTEMPTS; attempt++) {
            try {
                console.log(`Registry load attempt ${attempt}/${MAX_RETRY_ATTEMPTS}`);

                // Dynamic import of registry module
                const module = await import(REGISTRY_URL);

                // Extract default export (registry object)
                componentRegistry = module.default;

                if (!componentRegistry || typeof componentRegistry !== 'object') {
                    throw new Error('Invalid registry format: expected object with component templates');
                }

                console.log(`Component registry loaded successfully (${Object.keys(componentRegistry).length} components)`);
                registryLoadPromise = null;
                return componentRegistry;

            } catch (error) {
                lastError = error;
                console.warn(`Registry load attempt ${attempt} failed:`, error.message);

                // Don't delay after last attempt
                if (attempt < MAX_RETRY_ATTEMPTS) {
                    const delay = INITIAL_RETRY_DELAY * attempt; // Exponential backoff
                    console.log(`Retrying in ${delay}ms...`);
                    await new Promise(resolve => setTimeout(resolve, delay));
                }
            }
        }

        // All attempts failed
        registryLoadPromise = null;
        const errorMsg = `Failed to load component registry after ${MAX_RETRY_ATTEMPTS} attempts`;
        console.error(errorMsg, lastError);
        throw new Error(`${errorMsg}: ${lastError.message}`);
    })();

    return registryLoadPromise;
}

/**
 * Render dynamic component at runtime
 *
 * Pattern: Alpine.js Magic Function [Cognitive Load: 12]
 * - Registry loading: 3
 * - Component lookup: 2
 * - Template rendering: 3
 * - Error handling: 4
 *
 * This is the main entry point called by Alpine.js x-init directive.
 * It loads the registry (if needed), gets the template function, and renders
 * the component HTML into the target element.
 *
 * Error Handling Strategy (non-fatal, graceful degradation):
 * - Missing component: HTML comment warning (silent to user)
 * - Network error: Retry with backoff, show error if all fail
 * - Template error: Log to console, show error in dev mode
 *
 * @param {HTMLElement} el - Target element to render into
 * @param {string} componentName - Name of component to render (e.g., "Hero2436")
 * @param {Object} props - Props to pass to component template
 * @returns {Promise<void>}
 */
async function renderDynamicComponent(el, componentName, props) {
    try {
        console.log(`Rendering component: ${componentName}`, props);

        // STEP 1: Load registry if not loaded (COGNITIVE LOAD: 3)
        let registry;
        try {
            registry = await loadComponentRegistry();
        } catch (error) {
            // Network error - show error message
            const errorMsg = `Failed to load component registry: ${error.message}`;
            console.error(errorMsg);
            el.innerHTML = `<!-- Error: ${errorMsg} -->
                <div style="padding: 1rem; background: #fee; border-left: 4px solid #c00; color: #800;">
                    <strong>Component Error:</strong> Unable to load component registry.
                    Check console for details.
                </div>`;
            return;
        }

        // STEP 2: Get component template function (COGNITIVE LOAD: 2)
        // Try exact match first, then case-insensitive match
        let templateFn = registry[componentName];

        if (!templateFn) {
            // Try case-insensitive lookup
            // This handles cases where JSON has 'hero2436' but registry has 'Hero2436'
            const lowerName = componentName.toLowerCase();
            const registryKey = Object.keys(registry).find(key => key.toLowerCase() === lowerName);
            if (registryKey) {
                templateFn = registry[registryKey];
                console.log(`[Runtime Components] Case-insensitive match: '${componentName}' → '${registryKey}'`);
            }
        }

        if (!templateFn) {
            // Component not found - graceful degradation
            const warningMsg = `Component '${componentName}' not found in registry`;
            console.warn(warningMsg);
            console.log('Available components:', Object.keys(registry).join(', '));

            // Insert HTML comment (visible in DevTools, not to user)
            el.innerHTML = `<!-- Warning: ${warningMsg} -->`;
            return;
        }

        if (typeof templateFn !== 'function') {
            throw new Error(`Invalid component template for '${componentName}': expected function, got ${typeof templateFn}`);
        }

        // STEP 2.5: Normalize props - strip 'this.' prefix from keys
        // The server-side replaceVarRefsWithThis adds 'this.' prefix to field names
        // but component templates expect clean prop names (props.title not props['this.title'])
        const normalizedProps = {};
        if (props) {
            for (const [key, value] of Object.entries(props)) {
                const cleanKey = key.startsWith('this.') ? key.substring(5) : key;
                normalizedProps[cleanKey] = value;
            }
        }

        // STEP 3: Render component (COGNITIVE LOAD: 3)
        let html;
        try {
            // Call template function with normalized props
            html = templateFn(normalizedProps);
        } catch (error) {
            // Template error - show error in dev mode
            const errorMsg = `Template error in '${componentName}': ${error.message}`;
            console.error(errorMsg, error);
            el.innerHTML = `<!-- Error: ${errorMsg} -->
                <div style="padding: 1rem; background: #fee; border-left: 4px solid #c00; color: #800;">
                    <strong>Template Error in ${componentName}:</strong> ${error.message}
                </div>`;
            return;
        }

        if (typeof html !== 'string') {
            throw new Error(`Invalid template output for '${componentName}': expected string, got ${typeof html}`);
        }

        // STEP 4: Wrap component in x-data scope and set element content (COGNITIVE LOAD: 3)
        // Components from the registry may have x-for and other directives that reference props
        // We need to provide props in an Alpine.js x-data scope so Alpine can resolve them
        // IMPORTANT: Wrap props under a "props" key to match the template's "props.*" references
        // Example: x-for="item in props.todos" needs x-data to have { props: { todos: [...] } }
        const xDataScope = { props: normalizedProps };
        const xDataJSON = JSON.stringify(xDataScope)
            .replace(/</g, '\\u003c')  // Escape < for security (prevents XSS)
            .replace(/>/g, '\\u003e')  // Escape > for security
            .replace(/'/g, "\\'");      // Escape single quotes for attribute context

        el.innerHTML = `<div x-data='${xDataJSON}'>${html}</div>`;
        console.log(`Component '${componentName}' rendered successfully with x-data scope`);

        // STEP 5: Re-initialize Alpine directives in new content (COGNITIVE LOAD: 3)
        // This ensures any Alpine.js directives in the component template work
        if (window.Alpine) {
            // Use Alpine.js to initialize directives in the new content
            window.Alpine.initTree(el);
        }

    } catch (error) {
        // Unexpected error - log and show error message
        console.error(`Unexpected error rendering component '${componentName}':`, error);
        el.innerHTML = `<!-- Error: ${error.message} -->
            <div style="padding: 1rem; background: #fee; border-left: 4px solid #c00; color: #800;">
                <strong>Unexpected Error:</strong> ${error.message}
            </div>`;
    }
}

/**
 * Register Alpine.js magic function using plugin pattern
 *
 * CRITICAL FIX: Use Alpine.plugin() instead of direct Alpine.magic() call
 * This ensures proper registration timing and avoids race conditions
 *
 * Pattern: Alpine.js Plugin [Cognitive Load: 5]
 * - Plugin registration: 3
 * - Magic function definition: 2
 */
document.addEventListener('alpine:init', () => {
    console.log('[Runtime Components] Alpine.js initializing, registering plugin...');

    // Register as Alpine.js plugin (best practice for magic functions)
    window.Alpine.magic('renderDynamicComponent', () => {
        // Return the async render function
        return renderDynamicComponent;
    });

    console.log('[Runtime Components] ✓ $renderDynamicComponent magic registered successfully');
});

// Also try immediate registration if Alpine is already loaded (edge case)
if (window.Alpine && !window.Alpine.version) {
    // Alpine is loaded but not yet started - safe to register
    console.log('[Runtime Components] Alpine.js detected, registering magic immediately');
    window.Alpine.magic('renderDynamicComponent', () => {
        return renderDynamicComponent;
    });
}
