var response = http.do("GET", "https://icanhazip.com", [{
  "Content-Type": "text/html"
}], {});
console.log(response.body);
