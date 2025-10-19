/*
preload.js -> This file is executed before the HTML is loaded.
index.js -> This file is executed after the HTML is loaded.
*/

// Local JavaScript is only applied to the post in this folder
// It lets you have full control over any custom behavior for an individual post

// JavaScript is totally optional. Most blogs do not need it.
function hello() {
    alert("Hello from LOCAL index.js!");
}

console.log("Hello from LOCAL index.js!");