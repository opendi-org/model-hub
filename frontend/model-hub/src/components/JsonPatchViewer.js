import React from "react";
//legacy patch viewer, no longer in use
//some bugs are with some paths returned from the diff. 
const JsonPatchViewer = ({ lastVersionOfModel, commit }) => {

  if (commit['version'] === 0) {
    return <pre>{JSON.stringify(lastVersionOfModel, null, 2)}</pre>;
    
  }

  const patch = JSON.parse(commit.diff);
  let modifiedModel = JSON.parse(JSON.stringify(lastVersionOfModel));

  patch.forEach((operation) => {
    if (operation.op === "replace") {
      const pathSegments = operation.path.split("/").filter(Boolean);
      let obj = modifiedModel;
      for (let i = 0; i < pathSegments.length - 1; i++) {
        obj = obj[pathSegments[i]];
      }
      const key = pathSegments[pathSegments.length - 1];
      obj[key] = `<span style='background-color: lightblue'>${obj[key]}</span>`;
    }
    if (operation.op === "add") {
      const pathSegments = operation.path.split("/").filter(Boolean);
      let obj = modifiedModel;
      for (let i = 0; i < pathSegments.length - 1; i++) {
        obj = obj[pathSegments[i]];
      }
      const key = pathSegments[pathSegments.length - 1];
      obj[key] = `<span style='background-color: green'>${obj[key]}</span>`;
    }
    if (operation.op === "remove") {
      const pathSegments = operation.path.split("/").filter(Boolean);
      let obj = modifiedModel;
      for (let i = 0; i < pathSegments.length - 1; i++) {
        obj = obj[pathSegments[i]];
      }
      const key = pathSegments[pathSegments.length - 1];
      obj[key] = `<span style='background-color: red'>${obj[key]}</span>`;
    }

  
  });

  return (
    <div>
      <h3>Previous JSON Model</h3>
      <pre dangerouslySetInnerHTML={{ __html: JSON.stringify(modifiedModel, null, 2).replace(/\"/g, '') }} />
    </div>
  );
};

export default JsonPatchViewer;
