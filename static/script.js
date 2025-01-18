async function fetchCommands() {
  try {
    const response = await fetch("/commands");
    if (!response.ok) {
      throw new Error(`/commands request failed: ${response.status}`);
    }
    const commands = await response.json();
    commands.forEach(command => {
      const option = document.createElement("option");
      option.value = command;
      option.textContent = command;
      document.getElementById("select-command").appendChild(option);
    });
  } catch (error) {
    console.error("fetchCommands:", error);
  }
}

async function fetchStatus() {
  fetchCommands();
  try {
    const response = await fetch("/system-status");
    if (!response.ok) {
      throw new Error(`/system-status request failed: ${response.status}`);
    }
    const status = await response.json();
    document.getElementById("hostname").textContent = "Host: " + status.Hostname;
    document.getElementById("uptime").textContent = "Uptime: " + status.Uptime;
    fillDiskUsageTable(status);
  } catch (error) {
    console.error("fetch-status:", error);
    document.getElementById("status-output").textContent = "Fetching error";
  }
}

document.getElementById("execute-command").addEventListener("click", async () => {
  try {
    const command = document.getElementById("select-command").value;
    const response = await fetch(`/command/${command}`, {
      method: "PUT",
    });
    if (!response.ok) {
      throw new Error(`/command/${command} request failed: ${response.status}`);
    }
  } catch (error) {
    console.error("execute-command:", error);
  }
});

function fillDiskUsageTable(data) {
  const tableBody = document.getElementById("diskUsageTable").getElementsByTagName("tbody")[0];

  for (const [mountPoint, usage] of Object.entries(data.DisksUsage)) {
    const row = tableBody.insertRow();
    row.insertCell(0).textContent = mountPoint;
    row.insertCell(1).textContent = usage.Total;
    row.insertCell(2).textContent = usage.Used;
    row.insertCell(3).textContent = usage.Free;

    const useCell = row.insertCell(4);
    useCell.textContent = usage.UsedPercent;
    useCell.style.color = colorFromPercent(usage.UsedPercent);
  }
}

// 0% green -> 100% red
function colorFromPercent(percent) {
  percent = parseInt(percent);
  const red = Math.min(255, Math.floor((percent / 100) * 255));
  const green = 255 - red;
  return `rgb(${red}, ${green}, 0)`;
}

window.onload = fetchStatus;
