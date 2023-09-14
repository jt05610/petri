import cuid from "cuid";

console.log(cuid())
console.log(cuid())

// log current timestamp without timezone
console.log(new Date().toISOString().slice(0, 19).replace('T', ' '))