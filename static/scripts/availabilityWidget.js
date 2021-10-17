/***********************************************************/
/*           JAVASCRIPT FOR AVAILABILITY WIDGET            */
/***********************************************************/

/* function to define a comparison on two targetData objects */
function compareTargetData(a, b) {
    if (a.availability < b.availability) {
        return 1;
    } else if (a.availability > b.availability) {
        return -1;
    }
    return 0;
}

/* function to generate the availability score */
function generateAvailabilityScore(targetData) {
    let availabilityItem = document.createElement("div");
    availabilityItem.innerText = parseFloat(targetData.availability).toFixed(2);
    availabilityItem.className = "availability-item";

    return availabilityItem;
}

/* function to generate the availability bar */
function generateAvailabilityBar(targetData) {
    let availabilityBarItem = document.createElement("div");
    availabilityBarItem.className = "availability-bar-item";
    let availabilityBarContent = document.createElement("span");
    if (targetData.availability >= 0.8) {
        availabilityBarContent.className = "progress-bar progress-bar-green";
    } else if (targetData.availability >= 0.5) {
        availabilityBarContent.className = "progress-bar progress-bar-yellow";
    } else {
        availabilityBarContent.className = "progress-bar progress-bar-red";
    }
    availabilityBarContent.style.width = `${targetData.availability * 100}%`;
    availabilityBarItem.appendChild(availabilityBarContent);

    return availabilityBarItem;
}

/* function to generate the target hostname */
function generateAvailabilityHostname(targetData) {
    let availabilityHostname = document.createElement("div");
    availabilityHostname.innerText = targetData.target;
    availabilityHostname.className = "availability-hostname";

    return availabilityHostname;
}

/* function to refresh the availability table */
function updateAvailibilityTable(duration) {
    fetch("/metrics/availability?" + new URLSearchParams({duration: duration})).then(async (resp) => {
        const data = await resp.json();

        const widgetBody = document.getElementById("availability-widget-body");
        widgetBody.innerHTML = "";


        for (const targetData of data.sort(compareTargetData)) {
            // Availability score
            widgetBody.appendChild(generateAvailabilityScore(targetData));

            // Availability progress bar
            widgetBody.appendChild(generateAvailabilityBar(targetData));
            
            // Target
            widgetBody.appendChild(generateAvailabilityHostname(targetData));
        }

    })
}

/* Instantiate the availability table */
updateAvailibilityTable("-10m");

/* Refresh table periodically */
setInterval(() => {
    updateAvailibilityTable("-10m");
}, 10 * 1000) // every 10 sec
