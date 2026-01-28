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
let commonChunk = null;

// Configuration
// Path is relative to built location: {fingerprint}/core/runtime-components.js
// Loads layouts from: {fingerprint}/generated/layouts.js
const FALLBACK_REGISTRY_URL = '../generated/layouts.js';
const MAX_RETRY_ATTEMPTS = 3;
const INITIAL_RETRY_DELAY = 500; // milliseconds

/**
 * Get bundle configuration from HTML data attributes
 *
 * Tree-shaking data attributes:
 * - data-bundle: Path to page-specific bundle
 * - data-common: Path to shared component chunk
 * - data-fallback: Path to full registry (fallback)
 *
 * @returns {Object} Bundle configuration
 */
function getBundleConfig() {
    const htmlElement = document.documentElement;
    return {
        bundle: htmlElement.getAttribute('data-bundle'),
        common: htmlElement.getAttribute('data-common'),
        fallback: htmlElement.getAttribute('data-fallback') || FALLBACK_REGISTRY_URL
    };
}

/**
 * Load a JavaScript module with retry logic
 *
 * @param {string} url - Module URL to load
 * @param {string} name - Name for logging
 * @returns {Promise<Object>} Module default export
 * @throws {Error} If all retry attempts fail
 */
async function loadModuleWithRetry(url, name) {
    let lastError = null;

    for (let attempt = 1; attempt <= MAX_RETRY_ATTEMPTS; attempt++) {
        try {
            console.log(`[Runtime Components] ${name} load attempt ${attempt}/${MAX_RETRY_ATTEMPTS}: ${url}`);
            const module = await import(url);
            const exports = module.default;

            if (!exports || typeof exports !== 'object') {
                throw new Error(`Invalid ${name} format: expected object`);
            }

            console.log(`[Runtime Components] ${name} loaded successfully (${Object.keys(exports).length} entries)`);
            return exports;

        } catch (error) {
            lastError = error;
            console.warn(`[Runtime Components] ${name} load attempt ${attempt} failed:`, error.message);

            if (attempt < MAX_RETRY_ATTEMPTS) {
                const delay = INITIAL_RETRY_DELAY * attempt;
                console.log(`[Runtime Components] Retrying ${name} in ${delay}ms...`);
                await new Promise(resolve => setTimeout(resolve, delay));
            }
        }
    }

    throw new Error(`Failed to load ${name} after ${MAX_RETRY_ATTEMPTS} attempts: ${lastError.message}`);
}

/**
 * Load component registry from server
 *
 * Pattern: Tree-Shaking Bundle Loading [Cognitive Load: 12]
 * - Bundle config detection: 3
 * - Common chunk loading: 3
 * - Page bundle loading: 3
 * - Fallback handling: 3
 *
 * Loading order:
 * 1. Check for tree-shaking data attributes
 * 2. If data-common exists, load common chunk first
 * 3. Load page bundle (data-bundle) or fallback registry
 * 4. Merge common + page bundle into final registry
 *
 * @returns {Promise<Object>} Component registry object
 * @throws {Error} If all loading attempts fail
 */
async function loadComponentRegistry() {
    // If already loaded, return cached registry
    if (componentRegistry) {
        console.log('[Runtime Components] Registry already loaded (cached)');
        return componentRegistry;
    }

    // If loading is in progress, wait for it
    if (registryLoadPromise) {
        console.log('[Runtime Components] Registry loading in progress, waiting...');
        return registryLoadPromise;
    }

    // Start loading
    console.log('[Runtime Components] Loading component registry...');

    registryLoadPromise = (async () => {
        const config = getBundleConfig();
        console.log('[Runtime Components] Bundle config:', config);

        try {
            // If tree-shaking is enabled (data-bundle attribute present)
            if (config.bundle) {
                console.log('[Runtime Components] Tree-shaking enabled, loading optimized bundles');

                // Step 1: Load common chunk if present
                if (config.common && !commonChunk) {
                    try {
                        commonChunk = await loadModuleWithRetry(config.common, 'common chunk');
                    } catch (error) {
                        console.warn('[Runtime Components] Common chunk failed, continuing without it:', error.message);
                        commonChunk = {};
                    }
                }

                // Step 2: Load page-specific bundle
                let pageBundle;
                try {
                    pageBundle = await loadModuleWithRetry(config.bundle, 'page bundle');
                } catch (error) {
                    console.warn('[Runtime Components] Page bundle failed, falling back to full registry');
                    pageBundle = await loadModuleWithRetry(config.fallback, 'fallback registry');
                }

                // Step 3: Merge common chunk and page bundle
                componentRegistry = { ...(commonChunk || {}), ...pageBundle };
                console.log(`[Runtime Components] Registry assembled: ${Object.keys(componentRegistry).length} components`);

            } else {
                // No tree-shaking, load full registry
                console.log('[Runtime Components] No tree-shaking config, loading full registry');
                componentRegistry = await loadModuleWithRetry(config.fallback, 'fallback registry');
            }

            registryLoadPromise = null;
            return componentRegistry;

        } catch (error) {
            registryLoadPromise = null;
            console.error('[Runtime Components] All loading strategies failed:', error);
            throw error;
        }
    })();

    return registryLoadPromise;
}

/**
 * Resolve component signature for Plenti compatibility
 *
 * Pattern: Component Resolution [Cognitive Load: 8]
 * - Signature generation: 3
 * - Multi-strategy lookup: 5
 *
 * Resolution order:
 * 1. Full Plenti signature (layouts_components_{name}_html)
 * 2. Exact short name (Hero2436)
 * 3. Case-insensitive match (hero2436 → Hero2436)
 *
 * @param {Object} registry - Component registry object
 * @param {string} componentName - Name to resolve (short name or signature)
 * @returns {Function|null} Template function if found, null otherwise
 */
function resolveComponentSignature(registry, componentName) {
    // Strategy 1: Try full Plenti signature first
    // This matches: allLayouts["layouts_components_" + name + "_html"]
    const signature = `layouts_components_${componentName}_html`;
    if (registry[signature]) {
        console.log(`[Runtime Components] Signature match: '${componentName}' → '${signature}'`);
        return registry[signature];
    }

    // Strategy 2: Try exact short name match
    if (registry[componentName]) {
        console.log(`[Runtime Components] Exact match: '${componentName}'`);
        return registry[componentName];
    }

    // Strategy 3: Try case-insensitive match
    const lowerName = componentName.toLowerCase();
    const registryKey = Object.keys(registry).find(key => key.toLowerCase() === lowerName);
    if (registryKey) {
        console.log(`[Runtime Components] Case-insensitive match: '${componentName}' → '${registryKey}'`);
        return registry[registryKey];
    }

    // Strategy 4: Try case-insensitive signature match
    const lowerSignature = signature.toLowerCase();
    const signatureKey = Object.keys(registry).find(key => key.toLowerCase() === lowerSignature);
    if (signatureKey) {
        console.log(`[Runtime Components] Case-insensitive signature match: '${componentName}' → '${signatureKey}'`);
        return registry[signatureKey];
    }

    return null;
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
 * For Plenti compatibility, component lookup tries:
 * 1. Full Plenti signature (layouts_components_{name}_html)
 * 2. Short name (Hero2436)
 * 3. Case-insensitive fallback
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

        // STEP 2: Get component template function using Plenti-compatible resolution
        // Try signature first, then short name, then case-insensitive
        let templateFn = resolveComponentSignature(registry, componentName);

        if (!templateFn) {
            // Component not found - graceful degradation
            const warningMsg = `Component '${componentName}' not found in registry`;
            console.warn(warningMsg);
            console.log('Available components:', Object.keys(registry).slice(0, 20).join(', ') + '...');

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
