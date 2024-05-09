const StatusNotStarted = 0;
const StatusRunning = 1;
const StatusStopping = 2;
const StatusStopped = 3;

const ActionAdd = "add";
const ActionStart = "start";
const ActionInterrupt = "interrupt";
const ActionKill = "kill";
const ActionDel = "del";
const ActionRefresh = "refresh";
const ActionEnv = "env";
const ActionError = "error";

function newProc(name) {
  return {
    "name":name,
    "path":"",
    "args":[],
    "env":[],
    "dir":"",
    "status":StatusStopped,
    "error":"",
    "stderr":""
  };
}

const app = {
  data() {
    const url = new URL("/ws", document.location.origin);
    url.protocol = "ws";
    const ws = new WebSocket(url.toString());
    ws.onmessage = this.msgHandler;
    ws.onerror = this.errHandler;
    ws.onopen = () => this.refreshProcs();
    return {
      proc: newProc(""),
      procs: [],
      globalEnv: [],
      editing: false,

      /*
      StatusStopped: StatusStopped,
      StatusRunning: StatusRunning,
      StatusStopping: StatusStopping,
      */

      ws: ws
    };
  },

  methods: {
    doAction(action, proc) {
      this.editing = false;
      this.ws.send(JSON.stringify({"action":action,"process":proc}));
    },

    startProc(proc) {
      if (proc.name === "")
        return;
      this.doAction(ActionStart, proc);
    },

    interruptProc(proc) {
      this.doAction(ActionInterrupt, proc);
    },

    killProc(proc) {
      this.doAction(ActionKill, proc);
    },

    delProc(proc) {
      if (!confirm(`Delete process named ${proc.name}?`))
        return;
      this.doAction(ActionDel, proc);
    },

    refreshProcs() {
      this.doAction(ActionRefresh, {});
    },

    newProcess() {
      let name = prompt("New process name");
      if (name === null) {
        return;
      }
      name = name.trim();
      if (name === "") {
        alert("Must input name");
        return;
      }
      this.editing = true;
      this.proc = newProc(name);
    },

    viewProcess(proc) {
      this.editing = false;
      this.proc = proc;
    },

    editProcess(proc) {
      if (proc === undefined) {
        proc = this.proc;
      }
      this.editing = true;
      this.proc = Object.assign({}, proc);
    },

    cancelEdit() {
      this.editing = false;
      this.proc = newProc("");
    },

    cloneProcess(proc) {
      if (proc === undefined) {
        proc = this.proc;
      }
      let name = prompt("New process name");
      if (name === null) {
        return;
      }
      name = name.trim();
      if (name === "") {
        alert("Must input name");
        return;
      }
      this.proc = Object.assign({}, proc);
      this.proc.name = name;
      this.editing = true;
    },

    msgHandler(event) {
      const msg = JSON.parse(event.data);
      switch (msg.action) {
        case ActionAdd:
          this.procs.push(msg.process);
          this.procs.sort((a, b) => (a.name > b.name) ? 1 : -1);
          break;
        case ActionStart:
          var proc = this.procs.find(proc => proc.name == msg.process.name);
          if (proc !== undefined)
            Object.assign(proc, msg.process);
          else
            this.doAction(ActionRefresh, {});
          break;
        case ActionKill:
          var proc = this.procs.find(proc => proc.name == msg.process.name);
          if (proc !== undefined)
            Object.assign(proc, msg.process);
          else
            this.refreshProcs();
          break;
        case ActionDel:
          var index = this.procs.indexOf(proc => proc.name == msg.process.name);
          if (index > -1)
            this.procs.splice(index, 1);
          else
            this.refreshProcs();
          break;
        case ActionRefresh:
          this.procs = JSON.parse(msg.contents);
          this.procs.sort((a, b) => (a.name > b.name) ? 1 : -1);
          break;
        case ActionEnv:
          const env = JSON.parse(msg.contents);
          for (var kv of env) {
            const index = kv.indexOf("=");
            if (index === -1) {
              // Bad KV pair
              continue;
            }
            const key = kv.substring(0, index);
            const value = kv.substring(index + 1);
            this.globalEnv.push({"key": key, "value": value});
          }
          this.globalEnv.sort((a, b) => {
            if (a.key > b.key) {
              return 1;
            } else if (a.key < b.key) {
              return -1;
            } else if (a.value > b.value) {
              return 1;
            } else if (a.value < b.value) {
              return -1;
            }
            return 0;
          });
          break;
        case ActionError: alert(msg.contents); break;
        default: console.log(`received unexpected message: ${msg}`); break;
      }
    },

    errHandler(err) {
      console.log(`websocket error occurred: ${err}`);
    },

    statusToString(status) {
      switch (status) {
        case StatusNotStarted:
          return "NOT STARTED";
        case StatusRunning:
          return "RUNNING";
        case StatusStopping:
          return "STOPPING";
        case StatusStopped:
          return "STOPPED";
      }
      return "UNKNOWN";
    }
  }
}
