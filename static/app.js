const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateContent(results);
      });
    });
  },

  updateContent: (results) => {
    const content = document.getElementById("content");
    const blocks = [];

    results.forEach((result) => {
      const innerBlocks = [];
      let fragments = result.fragments
      let title = result.work_title
      fragments.forEach((fragment) => {
        innerBlocks.push(`<div class="card border-dark mb-3"><div class="card-body">${fragment}</div></div>`)
      })
      blocks.push(`<h3>${title}</h3>${innerBlocks.join("\n")}`)
    })
    content.innerHTML = blocks.join("\n");
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
