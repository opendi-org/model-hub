import React from "react";
import { JSONTree } from 'react-json-tree';

//note - we also could've used this built in library, but it doesn't follow the RFC 6902 standard for JSON patching, so we can't use it for our purposes.
//https://www.npmjs.com/package/react-json-view-compare

//Note: 

/*
[
  {
    "name": "Alice",
    "age": 30
  },
  {
    "name": "Bob",
    "age": 25
  }
]

Now, let's say you want to modify the age property of the first object in the array (index 0) to 35. The path for this operation would look like:


{
  "op": "replace",
  "path": "/0/age",
  "value": 35
}


*/

//helper method for seeing if a path exists

const JsonDiffViewer = ({ lastVersionOfModel, commit }) => {



    if (commit['version'] === 0) {
        if (lastVersionOfModel == null) {
          return <pre>No previous version</pre>
        }
        return <pre>{JSON.stringify(lastVersionOfModel, null, 2)}</pre>;
    }

    const patch = JSON.parse(commit.diff);
    let modifiedModel = JSON.parse(JSON.stringify(lastVersionOfModel));


    /*
    1. customLabelRenderer closes over pathsToColors
    pathsToColors is defined in the outer lexical scope.

    When customLabelRenderer is defined, JavaScript creates a closure — that is, a function that "remembers" the outer variables it needs.

    Even if customLabelRenderer is passed around or used later, it retains access to pathsToColors.

    2. Memory for pathsToColors stays alive
    As long as customLabelRenderer (or any other function that references pathsToColors) is reachable, JavaScript keeps pathsToColors in memory.

    This is intentional — the JS engine uses reachability as its core GC principle.


    */


    const pathsToColors = {}
    patch.forEach((operation) => {

        //console.log("operation.path", operation.path)

        if (operation.op === "replace") {
            pathsToColors[operation.path] = "lightblue";
          }
          if (operation.op === "add") {
            //if the patch is one where we're appending to the end of an array, we need to remove the hyphen from the path and set the highlighted path to the last index of the array
            if (operation.path[operation.path.length - 1] === "-") {
                const pathSegments = operation.path.split("/").filter(Boolean); //.filter(boolean) filter sout any empty strings
                //console.log("pathSegments", pathSegments)
                let obj = modifiedModel;
                //get to the  part of the object before the hyphen -- in other words, the array to be appended
                for (let i = 0; i < pathSegments.length - 1; i++) {
                  obj = obj[pathSegments[i]];
                }
                
                operation.path = operation.path.slice(0, -1) + (obj.length - 1);  // Removing the hyphen and adding '0'
                //console.log("operation.path", operation.path)
                pathsToColors[operation.path] = "pink";
            }
            else {
                pathsToColors[operation.path] = "green";
            }
            
          }
          if (operation.op === "remove") {
            pathsToColors[operation.path] = "red";
          }

    });

    const customLabelRenderer = (keys) => {

        
        
        const keyName = keys[0]; //copies keys[0]

        //keys are the keys of the current node in the JSON tree, but in reversed order
        const actualOrderedKeys = keys.slice(0, -1).reverse(); 
        //JavaScript will implicitly convert all elements to strings before joining them with the specified delimiter
        const path = '/' + actualOrderedKeys.join('/');
    

        //console.log("path", path)
        //console.log("keys", keys)

        if (path in pathsToColors) {
            const color = pathsToColors[path];
            return (
            <strong style={{ backgroundColor: color, borderRadius: '4px' }}>
                {keyName}
            </strong>
            );
        }

    
        return <span>{keyName}</span>;
    };
    //

    return (
        <JSONTree
            data={modifiedModel}
            labelRenderer={customLabelRenderer}
            shouldExpandNodeInitially={(keyName, data, level) => {
                // Log to check if the function is being called
                //console.log('shouldExpandNode called for:', keyName, data);
          
                return true; // Always expand
              }}
        />
    );
};

export default JsonDiffViewer;




