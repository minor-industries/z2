import {BumperControl} from "./bumper-control.js";

export function setupControls(
    containerId: string,
    suffix = ""
) {
    const bc1 = new BumperControl({
        containerId: containerId,
        label: 'Target Speed',
        variableName: `bike_target_speed${suffix}`,
        increment: 0.25,
        defaultValue: 35,
        fixed: 2
    });

    const bc2 = new BumperControl({
        containerId: containerId,
        label: 'Max Drift %',
        variableName: `bike_max_drift_pct${suffix}`,
        increment: 0.1,
        defaultValue: 2.0,
        fixed: 1
    });

    const bc3 = new BumperControl({
        containerId: containerId,
        label: 'Max Error %',
        variableName: `bike_allowed_error_pct${suffix}`,
        increment: 0.1,
        defaultValue: 1.0,
        fixed: 1
    });

    return [bc1, bc2, bc3];
}

export function createPresetControls() {
    ["A", "B", "C", "D"].forEach(v => {
        const containerId = `preset_${v}`;
        const suffix = `_${v}`;

        setupControls(containerId, suffix);

        new BumperControl({
            containerId: containerId,
            label: 'Timer (seconds)',
            variableName: `bike_preset_timer${suffix}`,
            increment: 10,
            defaultValue: 60 * 4,
            fixed: 0
        });
    });
}

export function registerPresets(controls: BumperControl[]) {
    ["A", "B", "C", "D"].forEach(v => {
        document.getElementById(`preset${v}`)!.addEventListener('click', (event) => {
            console.log("preset", v, controls);
        });
    });
}