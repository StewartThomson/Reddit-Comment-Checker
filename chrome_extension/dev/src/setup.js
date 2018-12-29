let editor = document.querySelector(".usertext-edit");
editor.appendChild(createElementFromHTML(`
<div class="checker">
    <button id="checkBtn" type="button">
        Check for duplicates
    </button>
</div>
`));

let editorParent = editor.parentElement;
editorParent.appendChild(createElementFromHTML(`
<div class="checker">
    <div id="resultContainer">
        <div class="spinner display-none">
          <div class="rect1"></div>
          <div class="rect2"></div>
          <div class="rect3"></div>
          <div class="rect4"></div>
          <div class="rect5"></div>
        </div>
    </div>
</div>
`));

let spinner = document.querySelector("#resultContainer .spinner");
let resultContainer = document.querySelector("#resultContainer");

document.getElementById("checkBtn").addEventListener("click", function (event) {
    let comment = event.path[2].childNodes[1].childNodes[0].value;
    if (comment) {
        spinner.classList.remove("display-none");
        let opts = {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                content: comment,
                path: location.pathname
            })
        };

        fetch('https://tv5chxc3fa.execute-api.us-east-2.amazonaws.com/default', opts).then(function (response) {
            return response.json();
        }).then(function (body) {
            body.sort((a, b) => {
                return b.similarity - a.similarity;
            });

            displaySimilarComments(body)
        }).catch(function (err) {
            console.error(err);
        }).finally(function() {
            spinner.classList.add("display-none");
        });
    }
});

function displaySimilarComments(commentsInfo) {
    if (!Array.isArray(commentsInfo) || commentsInfo.length <= 0) {
        resultContainer.innerHTML = `<p>No similar comments found.</p>`
        return;
    }

    let maxDisplay = 5;
    let count = 0;
    let markup = `<h3>Similar Comments:</h3>`;
    for (let commentInfo of commentsInfo) {
        if (count >= maxDisplay) {
            break;
        }
        let commentNode = document.querySelector(`#thing_t1_${commentInfo.comment.id}`).cloneNode(true);
        for (let child = commentNode.firstChild; child !== null; child = child.nextSibling) {
            //Remove replies
            if(child.classList.contains("child")) {
                child.remove();
                continue;
            }
            //Remove buggy buttons
            if(child.classList.contains("entry")) {
                for (let subchild = child.firstChild; subchild !== null; subchild = subchild.nextSibling) {
                    if(subchild.classList.contains("buttons")) {
                        for(let subsubchild = subchild.firstChild; subsubchild !== null; subsubchild = subsubchild.nextSibling) {
                            if(subsubchild.classList.length === 0 || subsubchild.classList.contains("reply-button")) {
                                let toRemove = subsubchild;
                                subsubchild = subsubchild.previousSibling;
                                toRemove.remove();
                            }
                        }
                    }
                }
            }
        }
        markup += `
            <button class="accordion" type="button">Similarity: ${(commentInfo.similarity * 100).toFixed(2)}%</button>
            <div class="panel">
        `;
        markup += commentNode.outerHTML;
        markup += `</div>`;
        count++;
    }

    resultContainer.innerHTML = markup;

    let acc = document.querySelectorAll("#resultContainer .accordion");

    /*Accordion from https://www.w3schools.com/howto/howto_js_accordion.asp */
    for (let i = 0; i < acc.length; i++) {
        acc[i].addEventListener("click", function() {
            /* Toggle between adding and removing the "active" class,
            to highlight the button that controls the panel */
            this.classList.toggle("active");

            /* Toggle between hiding and showing the active panel */
            let panel = this.nextElementSibling;
            if (panel.style.maxHeight){
                panel.style.maxHeight = null;
            } else {
                panel.style.maxHeight = panel.scrollHeight + "px";
            }
        });
    }
}