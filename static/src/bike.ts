import {setupAnalysis} from "./analysis";
import {Env} from "./env";

export function setupBikeAnalysis(env: Env, date: string) {
    setupAnalysis(env, {
        date: date,
        ref: "bike",
        seriesNames: [
            "bike_avg_speed_long | time-bin",
            "bike_avg_speed_short | time-bin",
            "bike_instant_speed_min | time-bin",
            "bike_instant_speed_max | time-bin",
            "bike_target_speed | time-bin",
        ],
        title: "Avg Speed",
        ylabel: "speed (km/h)",
    });
}


