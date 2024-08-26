import {BumperControl} from "./bumper-control.js";
import {ApiClient} from "./api_client";

export function setupControls(
    apiClient: ApiClient,
    containerId: string,
    kind: string,
    suffix = ""
): BumperControl[] {
    if (kind === "bike") {
        return setupBikeControls(apiClient, containerId, suffix);
    } else if (kind === "rower") {
        return setupRowerControls(apiClient, containerId, suffix);
    }

    throw new Error(`unknown kind: ${kind}`);
}

function setupBikeControls(
    apiClient: ApiClient,
    containerId: string,
    suffix: string,
): BumperControl[] {
    const bc1 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Target Speed',
        variableName: `bike_target_speed${suffix}`,
        increment: 0.25,
        defaultValue: 35,
        fixed: 2
    });

    const bc2 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Max Drift %',
        variableName: `bike_max_drift_pct${suffix}`,
        increment: 0.1,
        defaultValue: 2.0,
        fixed: 1
    });

    const bc3 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Max Error %',
        variableName: `bike_allowed_error_pct${suffix}`,
        increment: 0.1,
        defaultValue: 1.0,
        fixed: 1
    });

    return [bc1, bc2, bc3];
}

function setupRowerControls(
    apiClient: ApiClient,
    containerId: string,
    suffix: string,
): BumperControl[] {
    const bc1 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Target Power',
        variableName: `rower_target_power${suffix}`,
        increment: 2.5,
        defaultValue: 100,
        fixed: 1
    });

    const bc2 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Max Drift %',
        variableName: `rower_max_drift_pct${suffix}`,
        increment: 0.1,
        defaultValue: 5.0,
        fixed: 1
    });

    const bc3 = new BumperControl(apiClient, {
        containerId: containerId,
        label: 'Max Error %',
        variableName: `rower_allowed_error_pct${suffix}`,
        increment: 0.1,
        defaultValue: 2.0,
        fixed: 1
    });

    return [bc1, bc2, bc3];
}

export function createPresetControls(apiClient: ApiClient, kind: string): void {
    ["A", "B", "C", "D"].forEach(v => {
        const containerId = `preset_${v}`;
        const suffix = `_${v}`;

        setupControls(apiClient, containerId, kind, suffix);

        new BumperControl(apiClient, {
            containerId: containerId,
            label: 'Timer (seconds)',
            variableName: `${kind}_preset_timer${suffix}`,
            increment: 10,
            defaultValue: 60 * 4,
            fixed: 0
        });
    });
}

type VariableLists = Record<string, string[]>;

const variableLists: VariableLists = {
    bike: [
        'bike_target_speed',
        'bike_max_drift_pct',
        'bike_allowed_error_pct',
        'bike_preset_timer'
    ],
    rower: [
        'rower_target_power',
        'rower_max_drift_pct',
        'rower_allowed_error_pct',
        'rower_preset_timer'
    ]
}

export async function registerPresets(
    apiClient: ApiClient,
    controls: BumperControl[],
    kind: string,
): Promise<void> {
    ["A", "B", "C", "D"].forEach(v => {
        document.getElementById(`preset${v}`)!.addEventListener('click', async () => {
            const suffix = `_${v}`;

            const variables = variableLists[kind];

            const presetNames = variables.map(name => `${name}${suffix}`);
            const resp = await apiClient.ReadVariables({variables: presetNames});

            const controlVariables = variables.slice(0, -1);

            for (let i = 0; i < controlVariables.length; i++) {
                const name = controlVariables[i];
                const control = controls.find(c => c.getVariableName() === name);
                if (!control) {
                    console.log(name, "control not found");
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

            const timerVariable = resp.variables.slice(-1)[0];

            const display = document.querySelector('#timerDisplay')!;
            if (timerVariable.present && timerVariable.value > 0) {
                startTimer(timerVariable.value, display as HTMLElement);
            } else {
                display.textContent = "00:00";
                clearInterval(timerInterval);
            }
        });
    });
}

let timerInterval: number | undefined;

function startTimer(duration: number, display: HTMLElement): void {
    const endTime = Date.now() + duration * 1000;

    if (timerInterval !== undefined) {
        clearInterval(timerInterval);
    }

    timerInterval = window.setInterval(() => {
        const remainingTime = Math.floor((endTime - Date.now()) / 1000);
        const minutes = Math.floor(Math.abs(remainingTime) / 60);
        const seconds = Math.abs(remainingTime) % 60;

        const displayMinutes = minutes < 10 ? "0" + minutes : minutes.toString();
        const displaySeconds = seconds < 10 ? "0" + seconds : seconds.toString();

        display.textContent = (remainingTime < 0 ? "-" : "") + displayMinutes + ":" + displaySeconds;

        if (remainingTime <= -3600) {  // Stop updating after 1 hour into negative
            clearInterval(timerInterval);
        }
    }, 50);
}
