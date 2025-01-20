async function fetchDashboardData() {
  try {
    const response = await fetch("/all");
    if (!response.ok) {
      throw new Error(`/all request failed: ${response.status}`);
    }
    const data = await response.json();
    document.getElementById("hostname").textContent = "Host: " + data.Hostname;
    document.getElementById("uptime").textContent = "Uptime: " + data.Uptime;
    fillDiskUsageTable(data);
    data.Commands.forEach(command => {
      const option = document.createElement("option");
      option.value = command;
      option.textContent = command;
      document.getElementById("select-command").appendChild(option);
    });
  } catch (error) {
    console.error(error);
    document.getElementById("command-output").textContent = "Fetching error";
  }
}

document.getElementById("execute-command").addEventListener("click", async () => {
  try {
    const command = document.getElementById("select-command").value;
    const response = await fetch(`/command/${command}`, {
      method: "PUT",
    });
    document.getElementById("command-output").textContent = await response.text();
    if (!response.ok) {
      throw new Error(`/command/${command} request failed: ${response.status}`);
    }
  } catch (error) {
    console.error(error);
    if (error instanceof TypeError) {
      document.getElementById("command-output").textContent = error.message;
    }
  }
});

function fillDiskUsageTable(data) {
  const tableBody = document.getElementById("disk-usage").getElementsByTagName("tbody")[0];

  if (Object.keys(data.DisksUsage).length === 0) {
    document.getElementById("disk-usage").style.display = "none";
  } else {
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
}

// 0% green -> 100% red
function colorFromPercent(percent) {
  percent = parseInt(percent);
  const red = Math.min(255, Math.floor((percent / 100) * 255));
  const green = 255 - red;
  return `rgb(${red}, ${green}, 0)`;
}

window.onload = fetchDashboardData;
