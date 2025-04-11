let t = document.querySelector(".todos.plenti-oQ3wky");
let num = t.dataset.number;
fetch('https://jsonplaceholder.typicode.com/todos').then((response) => {
    return response.json();
}).then((json) => {
    t.innerHTML = json.slice(0, num).map((todo) => {
        return "<tr><td>" + todo.title + "</td><td>" + todo.completed + "</td></tr>";
    }).join('');
});