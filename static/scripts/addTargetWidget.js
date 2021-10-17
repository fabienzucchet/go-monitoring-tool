/***********************************************************/
/*            JAVASCRIPT FOR NEW TARGET WIDGET             */
/***********************************************************/


/* Validation of url input : check that submitted url starts with http:// or https:// */
function isValidHttpUrl(url) {
    const expression = /(^https:\/\/.*|^http:\/\/.*)/;
    const regex = new RegExp(expression);

    return (url.match(regex))
}

/* Form validation */
function validateForm(data) {
    
    //Url validation
    if (!isValidHttpUrl(data.url)) {
        updateStatusMessage("error", "Please provide a valid URL");
        return false

        //Check_interval validation
    } else if (data.collectioninterval <= 0) {
        updateStatusMessage("error", "Please provide a valid collection interval");
        return false
    }

    return true

}

/* Remove the success/error message of the form */
function removeStatusMessage() {
    const statusMessage = document.getElementById("form-message-wrapper");
    statusMessage.innerText = "";
    statusMessage.className = "message-wrapper";
}

/* Update the success/error message of the form */
function updateStatusMessage(status, message) {
    const statusMessage = document.getElementById("form-message-wrapper");
    statusMessage.innerText = message;
    statusMessage.className = `message-wrapper message-${status}`;
    setTimeout(() => removeStatusMessage(), 15 * 1000);
}

/* Handle form submit : validate and post data to create target */
async function handleFormSubmit(event) {
    event.preventDefault();

    const form = event.currentTarget;
    const url = form.action;

    try {
        const formData = new FormData(form);

        const plainFormData = Object.fromEntries(formData.entries());

        //Validation
        if (validateForm(plainFormData)) {
            const formDataJsonString = JSON.stringify(plainFormData);

            //Create website through an API call
            const fetchOptions = {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Accept: "application/json",
                },
                body: formDataJsonString,
            }

            const response = await fetch(url, fetchOptions);

            if (!response.ok) {
                const errorMessage = await response.text();
                updateStatusMessage("error", errorMessage);
            }
            const res = await response.json();
            
            // Display success/error message to the user
            updateStatusMessage(res.status, res.message);
        }

    } catch (error) {
        console.error(error);
    }
}

/* Bind the custom submit function to the form */
const addWebsiteForm = document.getElementById("new-target-form");
addWebsiteForm.addEventListener("submit", handleFormSubmit);

