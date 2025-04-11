
	let htmlEl = document.getElementsByTagName("html")[0];
	let cms_fields_str = htmlEl.getAttribute("x-data");
	let cms_fields = eval("(" + cms_fields_str + ")");

	function createInputs(obj, container) {
		for (const key in obj) {
			if (Object.hasOwnProperty.call(obj, key)) {
				const input = document.createElement('input');
				input.type = 'text';
				let attribute = "x-model";
				if (typeof obj[key] === 'number') {
					input.type = 'number';
					attribute += ".number";
				}
				input.id = key;
				input.name = key;
				input.placeholder = key;
				input.setAttribute(attribute, key);
				const label = document.createElement('label');
				label.htmlFor = key;
				label.textContent = key;
				container.appendChild(label);
				container.appendChild(input);
				container.appendChild(document.createElement('br'));
			}
		}
	}
	const cms = document.getElementById('load_cms');
	createInputs(cms_fields, cms);
let d = document.querySelector("div.plenti-F1r2Hw");
d.style.border = "1px solid red";
let ds = document.querySelectorAll("div.double");
ds.forEach((d) => {
    d.style.color = "green";
});let d = document.querySelector("div.plenti-beYQkL");
d.style.border = "1px solid red";
let ds = document.querySelectorAll("div.double");
ds.forEach((d) => {
    d.style.color = "green";
});let d = document.querySelector("div.plenti-biHgxZ");
d.style.border = "1px solid red";
let ds = document.querySelectorAll("div.double");
ds.forEach((d) => {
    d.style.color = "green";
});let t = document.querySelector(".todos.plenti-nBPZCk");
let num = t.dataset.number;
fetch('https://jsonplaceholder.typicode.com/todos').then((response) => {
    return response.json();
}).then((json) => {
    t.innerHTML = json.slice(0, num).map((todo) => {
        return "<tr><td>" + todo.title + "</td><td>" + todo.completed + "</td></tr>";
    }).join('');
});let t = document.querySelector(".todos.plenti-Ngey2R");
let num = t.dataset.number;
fetch('https://jsonplaceholder.typicode.com/todos').then((response) => {
    return response.json();
}).then((json) => {
    t.innerHTML = json.slice(0, num).map((todo) => {
        return "<tr><td>" + todo.title + "</td><td>" + todo.completed + "</td></tr>";
    }).join('');
});let t = document.querySelector(".todos.plenti-NMno16");
let num = t.dataset.number;
fetch('https://jsonplaceholder.typicode.com/todos').then((response) => {
    return response.json();
}).then((json) => {
    t.innerHTML = json.slice(0, num).map((todo) => {
        return "<tr><td>" + todo.title + "</td><td>" + todo.completed + "</td></tr>";
    }).join('');
});