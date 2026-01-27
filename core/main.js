/**
 * Plenti-Compatible Runtime Entry Point
 *
 * This module initializes the Plenti-compatible runtime for Alpine.js pages.
 * It reads content from generated files and sets up the page data.
 *
 * This file is ejectable - developers can customize it by copying to their project's core/ directory.
 *
 * Pattern: Runtime Initialization Module [Cognitive Load: 20]
 * - Content loading: 5
 * - Layouts loading: 5
 * - Page initialization: 5
 * - Alpine.js integration: 5
 *
 * Usage:
 *   This file is automatically included in built pages.
 *   Import paths are relative to the fingerprinted directory.
 */

// Import generated content and layouts
// These paths are relative to the fingerprinted directory when built
import allContent from '../generated/content.js';
import allLayouts from '../generated/layouts.js';

/**
 * Get content file path from HTML data attribute
 *
 * Plenti stores the content path in data-content-filepath on the <html> tag.
 * Example: <html data-content-filepath="content/pages/about.json">
 *
 * @returns {string|null} Content file path or null if not found
 */
function getContentFilePath() {
    const htmlElement = document.documentElement;
    return htmlElement.getAttribute('data-content-filepath');
}

/**
 * Find content for current page
 *
 * Searches allContent array for the entry matching the content file path.
 *
 * @param {string} filepath - Content file path (e.g., "content/pages/about.json")
 * @returns {Object|null} Content object or null if not found
 */
function findContent(filepath) {
    if (!filepath || !allContent) {
        return null;
    }

    // Find content matching the filepath
    const content = allContent.find(c => c.filepath === filepath);
    if (content) {
        console.log(`[Plenti Core] Found content for: ${filepath}`);
        return content;
    }

    // Try matching by path (without extension)
    const pathWithoutExt = filepath.replace(/\.json$/, '').replace(/^content\//, '');
    const contentByPath = allContent.find(c => c.path === pathWithoutExt);
    if (contentByPath) {
        console.log(`[Plenti Core] Found content by path: ${pathWithoutExt}`);
        return contentByPath;
    }

    console.warn(`[Plenti Core] Content not found for: ${filepath}`);
    return null;
}

/**
 * Render component by name
 *
 * This is the Plenti-compatible $component magic helper.
 * It looks up a component in allLayouts and renders it with props.
 *
 * Usage in Alpine.js:
 *   <div x-html="$component('Hero2436', {title: 'Hello'})"></div>
 *
 * @param {string} name - Component name (e.g., "Hero2436")
 * @param {Object} props - Props to pass to component
 * @returns {string} Rendered HTML or error message
 */
function renderComponent(name, props = {}) {
    if (!allLayouts) {
        console.error(`[Plenti Core] Layouts not loaded`);
        return `<!-- Error: Layouts not loaded -->`;
    }

    // Try Plenti signature first
    const signature = `layouts_components_${name}_html`;
    let templateFn = allLayouts[signature];

    // Fall back to short name
    if (!templateFn) {
        templateFn = allLayouts[name];
    }

    // Try case-insensitive match
    if (!templateFn) {
        const lowerName = name.toLowerCase();
        const key = Object.keys(allLayouts).find(k => k.toLowerCase() === lowerName);
        if (key) {
            templateFn = allLayouts[key];
        }
    }

    if (!templateFn) {
        console.warn(`[Plenti Core] Component not found: ${name}`);
        return `<!-- Component not found: ${name} -->`;
    }

    if (typeof templateFn !== 'function') {
        console.error(`[Plenti Core] Invalid component: ${name}`);
        return `<!-- Invalid component: ${name} -->`;
    }

    try {
        return templateFn(props);
    } catch (error) {
        console.error(`[Plenti Core] Error rendering ${name}:`, error);
        return `<!-- Error rendering ${name}: ${error.message} -->`;
    }
}

/**
 * Initialize Plenti runtime
 *
 * Sets up Alpine.js stores and magic helpers for Plenti compatibility.
 */
function initPlenti() {
    console.log('[Plenti Core] Initializing...');

    // Get content file path from HTML
    const contentFilePath = getContentFilePath();
    console.log(`[Plenti Core] Content path: ${contentFilePath}`);

    // Find content for current page
    const content = findContent(contentFilePath);

    // Export to window for debugging
    window.allContent = allContent;
    window.allLayouts = allLayouts;
    window.content = content;

    console.log(`[Plenti Core] Content loaded:`, content ? 'yes' : 'no');
    console.log(`[Plenti Core] ${allContent?.length || 0} content items`);
    console.log(`[Plenti Core] ${Object.keys(allLayouts || {}).length} layout components`);

    // Wait for Alpine.js to be ready
    document.addEventListener('alpine:init', () => {
        console.log('[Plenti Core] Alpine.js initializing...');

        // Register Plenti store with content data
        window.Alpine.store('plenti', {
            content: content,
            allContent: allContent,
            allLayouts: allLayouts,

            // Get content field value
            field(name) {
                return this.content?.fields?.[name] || null;
            },

            // Get component by name
            getComponent(name) {
                return allLayouts?.[name] || allLayouts?.[`layouts_components_${name}_html`];
            }
        });

        // Register $component magic helper
        window.Alpine.magic('component', () => {
            return renderComponent;
        });

        console.log('[Plenti Core] ✓ Plenti store and $component magic registered');
    });
}

// Initialize on DOMContentLoaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initPlenti);
} else {
    initPlenti();
}

// Export for use in other modules
export { allContent, allLayouts, getContentFilePath, findContent, renderComponent };
