import {BumperControl} from "./bumper-control.js";
import {ReadVariables, ReadVariablesResp} from "./api.js";

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

export async function registerPresets(controls: BumperControl[]) {
    ["A", "B", "C", "D"].forEach(v => {
        document.getElementById(`preset${v}`)!.addEventListener('click', async (event) => {
            const suffix = `_${v}`;

            const variables = [
                'bike_target_speed',
                'bike_max_drift_pct',
                'bike_allowed_error_pct'
            ];

            const presetNames = variables.map(name => `${name}${suffix}`);
            const resp: ReadVariablesResp = await ReadVariables({variables: presetNames});

            for (let i = 0; i < variables.length; i++) {
                const name = variables[i];
                const control = controls.find(c => c.getVariableName() === name);
                if (!control) {
                    console.log(name, "control not found")
                    continue;
                }
                const preset = resp.variables[i];
                if (!preset.present) {
                    console.log(name, "preset not present")
                    continue;
                }
                await control.setValue(preset.value);
                console.log("Preset loaded:", name, preset.value);
            }
        });
    });
}
