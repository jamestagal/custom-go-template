// Auto-generated bundle for pages/jim-test
// Unique components: jim_test_advanced_loops, jim_test_age_examples, jim_test_animals_loop, jim_test_greeting, jim_test_notifications, jim_test_todos, jim_test_user_profiles
// Imports common chunk with: footer, head, header, html, nav

import common from '../../common.5fe3ef5b.js';

const unique = {
  'jim_test_advanced_loops': (props) => `<div style="background-color: #f9fafb; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; margin: 2rem 0;">
  <h2>Advanced Loop Patterns</h2>

  <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Array Spread in Loop:</h3>
  <div style="background-color: white; padding: 1rem; border-radius: 0.375rem; margin-bottom: 1.5rem;"><template x-for="animal in props.["🦄 unicorn", ...animals]">
      <div style="padding: 0.25rem 0; color: #1f2937;"><span x-text="animal"></span></div></template>
  </div>

  <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Inline Array Iteration:</h3>
  <div style="background-color: white; padding: 1rem; border-radius: 0.375rem;"><template x-for="word in props.["Waller", "loves Plenti", "uses AI", "is Australian"]">
      <div style="padding: 0.25rem 0; color: #1f2937;">${props.name} <span x-text="word"></span></div></template><template x-for="phrase in props.["rocks", "codes", "innovates"]">
      <div style="padding: 0.25rem 0; color: #1f2937; font-style: italic;">${props.name} <span x-text="phrase"></span>!</div></template>
  </div>
</div><style>
  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  h3 {
    color: #374151;
    font-size: 1.125rem;
    margin-bottom: 0.5rem;
    font-weight: 500;
  }
</style>`,
  'jim_test_age_examples': (props) => `<div style="margin: 2rem 0;">
  <h2>Age Component Examples</h2>
  <div style="margin-bottom: 0.5rem; font-size: 0.875rem; color: #666;">Passing dynamic props to components:</div>
</div><style>
  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }
</style>`,
  'jim_test_animals_loop': (props) => `<div class="animals" style="background-color: #f0f0f0; padding: 1rem; margin: 2rem 0;">
  <h2>Animals Loop with Advanced Features</h2><template x-for="animal in props.animals"><template x-if="animal == \"cat\"">
      <div>Hi <span x-text="animal"></span>!</div></template><template x-else>
      <div>Bye <span x-text="animal"></span>.</div></template>
    <div :class="animal">${props.name} likes: <span x-text="animal"></span>s</div>
    <div style="color: #666; font-size: 0.875rem;">Backwards: <span x-text="animal.split('').reverse().join('')"></span></div>
    <button :onclick="animals = animals.filter(a => a !== animal)">Remove <span x-text="animal"></span></button>
    <br></br></template>
  <h3>Add new animal:</h3>
  <input type="text" name="newAnimal" x-model="newAnimal" placeholder="animal name"></input>
  <button onclick="{animals = [newAnimal, ...animals]; newAnimal = ''}">Add Animal</button>
</div><style>
  .animals {
    border-radius: 0.5rem;
  }
  .animals div {
    margin: 0.5rem 0;
  }

  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  h3 {
    color: #374151;
    font-size: 1.125rem;
    margin-bottom: 0.5rem;
    font-weight: 500;
  }

  /* Button styling */
  button {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
    border: none;
    outline: none;
  }

  .animals button {
    background-color: #dc2626;
    color: white;
    padding: 0.375rem 0.75rem;
    font-size: 0.8rem;
  }

  .animals button:hover {
    background-color: #b91c1c;
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .animals button:active {
    transform: translateY(0);
    box-shadow: none;
  }

  .animals > button {
    background-color: #16a34a;
    color: white;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    margin-top: 0.5rem;
  }

  .animals > button:hover {
    background-color: #15803d;
  }

  .animals input {
    padding: 0.5rem;
    border: 2px solid #e5e7eb;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    margin-right: 0.5rem;
    outline: none;
    transition: border-color 0.2s ease;
  }

  .animals input:focus {
    border-color: #16a34a;
  }

  .animals h3 {
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;
    color: #1f2937;
    font-size: 1.125rem;
  }
</style>`,
  'jim_test_greeting': (props) => `<div class="text-content-container">
  <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
    ⚡ Build time: <strong>${props.buildTime}</strong>
  </div>

  <h1>${props.salutation} ${props.name}!</h1>

  <!-- Basic conditionals --><template x-if="name.length > 3">
    <div id="praise">${props.name} is a long name</div><template x-if="age > 1">
      <div>Has been born</div></template></template><template x-else-if="name.length == 2">
    <div id="praise">${props.name} is medium</div></template><template x-else>
    <div id="praise">${props.name} is a short name</div></template>
</div><style>
  .text-content-container {
    max-width: 80rem;
    margin: 0 auto;
    /* padding: 10rem 2rem 2rem 2rem; */
  }
  h1 {
    color: orange;
  }
  #praise {
    font-size: 2rem;
    color: green;
  }
</style>`,
  'jim_test_notifications': (props) => `<div style="margin: 2rem 0;">
  <h2>Interactive Notification Examples</h2>
  <div style="display: flex; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap;"><template x-for="notif in props.notifications">
      <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #3b82f6; color: white;" :onclick="currentNotification = notif">
        Show <span x-text="notif.type"></span>
      </button></template>
    <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #6b7280; color: white;" onclick="{currentNotification = null}">
      Clear
    </button>
  </div><template x-if="currentNotification"></template>
</div><style>
  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  button {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
    border: none;
    outline: none;
  }

  button:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  button:active {
    transform: translateY(0);
    box-shadow: none;
  }
</style>`,
  'jim_test_todos': (props) => `<div style="margin: 2rem 0;">
  <h2>Task List - First 5 Tasks</h2>

  <h2 style="margin-top: 2rem;">Task List - Tasks 6-14</h2>
</div><style>
  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  /* Todos table styling */
  table {
    width: 100%;
    border-collapse: collapse;
    margin: 1rem 0;
    background-color: white;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  thead {
    background-color: #f3f4f6;
  }

  th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-weight: 600;
    color: #374151;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  td {
    padding: 0.75rem 1rem;
    border-top: 1px solid #e5e7eb;
    color: #1f2937;
  }

  tbody tr:hover {
    background-color: #f9fafb;
  }
</style>`,
  'jim_test_user_profiles': (props) => `<div style="margin: 2rem 0;">
  <h2>User Profile Examples with Object Props</h2>
</div><style>
  h2 {
    color: #1f2937;
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  /* Role badge color classes */
  .bg-red-100 {
    background-color: #fee2e2;
  }
  .text-red-800 {
    color: #991b1b;
  }
  .bg-purple-100 {
    background-color: #f3e8ff;
  }
  .text-purple-800 {
    color: #6b21a8;
  }
  .bg-blue-100 {
    background-color: #dbeafe;
  }
  .text-blue-800 {
    color: #1e40af;
  }
  .bg-green-100 {
    background-color: #dcfce7;
  }
  .text-green-800 {
    color: #166534;
  }
  .bg-gray-100 {
    background-color: #f3f4f6;
  }
  .text-gray-800 {
    color: #1f2937;
  }

  /* Role badge styling */
  .profile-role {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
  }
</style>`
};

export default { ...common, ...unique };
