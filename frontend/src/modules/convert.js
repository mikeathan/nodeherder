
// returns only exposes that have features (properties)
export function getFeatureExposes(device) {

    // find exposes with properties
    // let all = items.filter(item=> item.age==='18')
    //     return deviceimport 
    // });


    var list = []
    for (const [key, expose] of Object.entries(device.exposes)) {
        if (expose.properties != undefined) {
            list.push(expose)
        }
    }

    return list
}

export function getDeviceExposeNames(device) {

    var list = []
    for (const [key, expose] of Object.entries(device.exposes)) {
        list.push(expose.name)
    }

    return list
}
export function getDeviceExposeNamesMap(device) {

    var list = {}
    for (const [key, expose] of Object.entries(device.exposes)) {
        list[expose.name] = expose.name
    }
    return list
}


// returns only devices that have features (properties)
export function getFeatureDevices(devices) {

    var list = []
    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.properties != undefined) {
                list.push(device)
                break;
            }
        }
    }

    return list;
}



export function createMapFromObject(obj, key, value) {

    return Object.assign({}, ...obj.map((x) => (
        {
            [getMapValue(x, key)]: getMapValue(x, value)
        }
    )));
}
function getMapValue(obj, key) {
    if (obj.hasOwnProperty(key))
        return obj[key];
    throw new Error("Invalid map key.");
}