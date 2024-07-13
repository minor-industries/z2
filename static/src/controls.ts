import {BumperControl} from "./bumper-control.js";

export async function setupControls() {
    const bc1 = new BumperControl({
        containerId: 'controls-container',
        label: 'Target Speed',
        variableName: `bike_target_speed`,
        increment: 0.25,
        defaultValue: 35,
        fixed: 2
    });

    const bc2 = new BumperControl({
        containerId: 'controls-container',
        label: 'Max Drift %',
        variableName: `bike_max_drift_pct`,
        increment: 0.1,
        defaultValue: 2.0,
        fixed: 1
    });

    const bc3 = new BumperControl({
        containerId: 'controls-container',
        label: 'Max Error %',
        variableName: `bike_allowed_error_pct`,
        increment: 0.1,
        defaultValue: 1.0,
        fixed: 1
    });
}


