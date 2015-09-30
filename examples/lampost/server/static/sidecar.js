var joust = jupyterJsOutputArea;
var OutputModel = joust.OutputModel, OutputView = joust.OutputView;

var SideCar = function Sidecar(container, document) {
  this.document = document;
  this.container = container;

  // parentID -> OutputArea
  this.areas = new Map();
};

SideCar.prototype.consume = function consume (message) {
  if (! message.parent_header && ! message.parent_header.msg_id) {
      return;
  }
  if (message.header &&
        (message.header.msg_type === "status" || message.header.msg_type === "execute_input")
      )  {
      return;
  }

  var parentID = message.parent_header.msg_id;
  var area = this.areas[parentID];

  if(!area) {
    // Create it
    area = new OutputArea(this.document);
    area.el.id = parentID; // For later bookkeeping
    this.container.appendChild(area.el);

    // Keep a running tally of output areas
    this.areas[parentID] = area;
  }

  var consumed = area.consume(message);
  if (consumed) {
    area.el.scrollIntoView();
  }

  return consumed;

};

var OutputArea = function OutputArea(document) {
  this.model = new OutputModel();
  this.view = new OutputView(this.model, document);

  this.el = this.view.el;

  this.el.className = 'output-area';
};

OutputArea.prototype.consume = function consume(message) {
  return this.model.consumeMessage(message);
};

window.SideCar = SideCar;
