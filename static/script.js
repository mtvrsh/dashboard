function fetchDashboardData() {
  fetch("/all").then(response => {
    if (!response.ok) {
      throw new Error(`/all request failed: ${response.status}`);
    }
    return response.json();
  }).then(data => {
    document.getElementById("hostname").textContent = "Host: " + data.Hostname;
    document.getElementById("uptime").textContent = "Uptime: " + data.Uptime;
    fillDiskUsageTable(data);
    createButtons(data.Commands);
  }).catch(error => {
    console.error(error);
    document.getElementById("command-output").textContent = "Fetching error";
  });
}

function createButtons(commands) {
  commands.forEach(command => {
    const button = document.createElement("button");
    button.textContent = command;
    button.classList.add("command");
    button.onclick = function() {
      executeCommand(command);
    };
    document.getElementById("commands").appendChild(button);
  });
}

function executeCommand(command) {
  fetch(`/command/${command}`, {
    method: "PUT",
  }).then(response => {
    response.text().then(output => document.getElementById("command-output").textContent = output);
    if (!response.ok) {
      throw new Error(`/command/${command} request failed: ${response.status}`);
    }
  }).catch(error => {
    console.error(error);
    if (error instanceof TypeError) {
      document.getElementById("command-output").textContent = error.message;
    }
  });
}

function fillDiskUsageTable(data) {
  const tableBody = document.getElementById("disk-usage").getElementsByTagName("tbody")[0];

  if (Object.keys(data.DisksUsage).length === 0) {
    document.getElementById("disk-usage").style.display = "none";
  } else {
    Object.entries(data.DisksUsage).forEach(([mountedOn, diskUsage]) => {
      const row = tableBody.insertRow();
      row.insertCell(0).textContent = mountedOn;
      row.insertCell(1).textContent = diskUsage.Size;
      row.insertCell(2).textContent = diskUsage.Used;
      row.insertCell(3).textContent = diskUsage.Avail;

      const useCell = row.insertCell(4);
      useCell.textContent = diskUsage.UsePercent;
      useCell.style.color = colorFromPercent(diskUsage.UsePercent);
    });
  }
}

// 0% green -> 100% red
function colorFromPercent(percent) {
  percent = parseInt(percent);
  const red = Math.min(255, Math.floor((percent / 100) * 255));
  const green = 255 - red;
  return `rgb(${red}, ${green}, 0)`;
}

fetchDashboardData();
