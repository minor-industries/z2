import {setupAnalysis} from "./analysis";
import {Env} from "./env";

export function setupRowerAnalysis(env: Env, date: string) {
    setupAnalysis(env, {
        date: date,
        ref: "rower",
        seriesNames: [
            "rower_avg_power_long | time-bin",
            "rower_avg_power_short | time-bin",
            "rower_target_power | time-bin",
            "rower_power_min | time-bin",
            "rower_power_max | time-bin",
        ],
        title: "Avg Power",
        ylabel: "power (watts)",
    });
}


