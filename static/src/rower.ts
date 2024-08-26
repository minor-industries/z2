import {setupAnalysis} from "./analysis";
import {ApiClient} from "./api_client";

export function setupRowerAnalysis(apiClient: ApiClient, date: string) {
    setupAnalysis(apiClient, {
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


