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

document.getElementById("fetch-status").addEventListener("click", async () => {
  try {
    const response = await fetch("/system-status");
    if (!response.ok) {
      throw new Error(`/system-status request failed: ${response.status}`);
    }
    const data = await response.json();
    document.getElementById("status-output").textContent = JSON.stringify(data, null, 2);
  } catch (error) {
    console.error("fetch-status:", error);
    document.getElementById("status-output").textContent = "Fetching error";
  }
});

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

window.onload = fetchCommands;
