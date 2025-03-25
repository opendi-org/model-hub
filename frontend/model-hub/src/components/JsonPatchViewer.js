import React from "react";
const JsonPatchViewer = ({ lastVersionOfModel, commit }) => {
  const patch = JSON.parse(commit.diff);
  let modifiedModel = JSON.parse(JSON.stringify(lastVersionOfModel));

  patch.forEach((operation) => {
    if (operation.op !== "test") {
      const pathSegments = operation.path.split("/").filter(Boolean);
      let obj = modifiedModel;
      for (let i = 0; i < pathSegments.length - 1; i++) {
        obj = obj[pathSegments[i]];
      }
      const key = pathSegments[pathSegments.length - 1];
      obj[key] = `<span style='background-color: yellow'>${obj[key]}</span>`;
    }
  });

  return (
    <div>
      <h3>Modified JSON Model</h3>
      <pre dangerouslySetInnerHTML={{ __html: JSON.stringify(modifiedModel, null, 2).replace(/\"/g, '') }} />
    </div>
  );
};

export default JsonPatchViewer;
