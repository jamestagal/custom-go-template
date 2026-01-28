// Get the content JSON file path from the html tag
// Plenti adds data-content-filepath attribute during rendering
function getContentFilePath() {
    const htmlEl = document.documentElement;
    return htmlEl.getAttribute('data-content-filepath');
}

// Fetch content JSON from server
async function fetchContentJSON(filepath) {
    if (!filepath) return null;

    try {
        // Fetch the JSON content file
        const response = await fetch('/' + filepath);
        if (!response.ok) {
            console.warn('[CMS] Failed to fetch content:', response.status);
            return null;
        }
        return await response.json();
    } catch (e) {
        console.warn('[CMS] Error fetching content JSON:', e);
        return null;
    }
}

// Find the main content x-data on the page (fallback if no JSON)
function findMainXData() {
    // Look specifically within .content-wrapper for content x-data
    const contentWrapper = document.querySelector('.content-wrapper');
    if (contentWrapper) {
        const elementsWithXData = contentWrapper.querySelectorAll('[x-data]');
        for (const el of elementsWithXData) {
            const data = el.getAttribute("x-data");
            if (data && data !== "{}" && data.includes(':')) {
                return { element: el, data: data };
            }
        }
    }
    return null;
}

function createInputs(obj, container, targetElement) {
    for (const key in obj) {
        if (Object.hasOwnProperty.call(obj, key)) {
            const value = obj[key];

            // Skip functions
            if (typeof value === 'function') {
                continue;
            }

            const wrapper = document.createElement('div');
            wrapper.style.marginBottom = '12px';

            const label = document.createElement('label');
            label.htmlFor = key;
            label.textContent = key;
            wrapper.appendChild(label);
            wrapper.appendChild(document.createElement('br'));

            // Handle arrays
            if (Array.isArray(value)) {
                // For simple arrays (strings/numbers), show as comma-separated
                if (value.length === 0 || typeof value[0] !== 'object') {
                    const input = document.createElement('input');
                    input.type = 'text';
                    input.id = key;
                    input.name = key;
                    input.placeholder = key + ' (comma-separated)';
                    input.value = value.join(', ');
                    input.dataset.targetXData = 'true';
                    input.dataset.isArray = 'true';

                    input.addEventListener('input', function() {
                        if (targetElement && targetElement._x_dataStack) {
                            const alpineData = targetElement._x_dataStack[0];
                            if (alpineData && key in alpineData) {
                                alpineData[key] = input.value.split(',').map(s => s.trim()).filter(s => s);
                            }
                        }
                    });

                    wrapper.appendChild(input);
                } else {
                    // Complex array - show as JSON (read-only for now)
                    const textarea = document.createElement('textarea');
                    textarea.id = key;
                    textarea.name = key;
                    textarea.value = JSON.stringify(value, null, 2);
                    textarea.style.width = '100%';
                    textarea.style.minHeight = '60px';
                    textarea.style.fontSize = '11px';
                    textarea.style.fontFamily = 'monospace';
                    wrapper.appendChild(textarea);
                }
            }
            // Handle objects (non-null, non-array)
            else if (typeof value === 'object' && value !== null) {
                const textarea = document.createElement('textarea');
                textarea.id = key;
                textarea.name = key;
                textarea.value = JSON.stringify(value, null, 2);
                textarea.style.width = '100%';
                textarea.style.minHeight = '60px';
                textarea.style.fontSize = '11px';
                textarea.style.fontFamily = 'monospace';
                wrapper.appendChild(textarea);
            }
            // Handle primitives (string, number, boolean)
            else {
                const input = document.createElement('input');
                input.type = 'text';
                if (typeof value === 'number') {
                    input.type = 'number';
                }
                input.id = key;
                input.name = key;
                input.placeholder = key;
                input.value = value;
                input.dataset.targetXData = 'true';

                input.addEventListener('input', function() {
                    if (targetElement && targetElement._x_dataStack) {
                        const alpineData = targetElement._x_dataStack[0];
                        if (alpineData && key in alpineData) {
                            alpineData[key] = input.type === 'number' ? Number(input.value) : input.value;
                        }
                    }
                });

                wrapper.appendChild(input);
            }

            container.appendChild(wrapper);
        }
    }
}

// Initialize CMS - async to allow fetching JSON
async function initCMS() {
    const cms = document.getElementById('plenti_cms');
    if (!cms) {
        console.warn('[CMS] #plenti_cms element not found in DOM - CMS UI disabled');
        return;
    }

    // Try to load from content JSON first (the Plenti way)
    const contentFilePath = getContentFilePath();
    if (contentFilePath) {
        console.log('[CMS] Found content filepath:', contentFilePath);
        const contentJSON = await fetchContentJSON(contentFilePath);

        if (contentJSON) {
            // Add header showing the file path
            const header = document.createElement('div');
            header.style.cssText = 'font-size: 11px; color: #666; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid #ddd;';
            header.innerHTML = `<strong>Editing:</strong><br>${contentFilePath}`;
            cms.appendChild(header);

            // For Plenti component-based pages, show component fields
            if (contentJSON.components && Array.isArray(contentJSON.components)) {
                contentJSON.components.forEach((comp, index) => {
                    const compHeader = document.createElement('h4');
                    compHeader.textContent = comp.name || `Component ${index + 1}`;
                    compHeader.style.cssText = 'margin: 15px 0 10px; padding-top: 10px; border-top: 1px solid #ddd; color: #333;';
                    cms.appendChild(compHeader);

                    if (comp.fields) {
                        createInputs(comp.fields, cms, null);
                    }
                });
            } else {
                // For flat JSON structure, show all fields directly
                createInputs(contentJSON, cms, null);
            }

            console.log('[CMS] Loaded content from JSON:', contentFilePath);
            return;
        }
    }

    // Fallback: try to find x-data on the page
    const xDataInfo = findMainXData();
    if (xDataInfo) {
        try {
            const cms_fields = eval("(" + xDataInfo.data + ")");
            createInputs(cms_fields, cms, xDataInfo.element);
            console.log('[CMS] Fallback: Found x-data on element:', xDataInfo.element.tagName);
        } catch (e) {
            console.warn('[CMS] Failed to parse x-data:', e);
        }
    } else {
        const msg = document.createElement('p');
        msg.textContent = 'No editable content found on this page.';
        msg.style.color = '#666';
        cms.appendChild(msg);
        console.warn('[CMS] No content JSON or x-data found');
    }
}

// Run CMS initialization
initCMS();

// Toggle button handler
const toggleButton = document.getElementById('toggle_plenti_cms');
if (toggleButton) {
    toggleButton.addEventListener('click', function () {
        const menu = document.getElementById('plenti_cms');
        if (menu) {
            menu.classList.toggle('menu-visible');
        }
    });
} else {
    console.warn('[CMS] #toggle_plenti_cms button not found - CMS toggle disabled');
}
