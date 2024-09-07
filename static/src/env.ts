import {ApiClient, DefaultApiClient} from "./api_client";
import * as calendar from "calendar";

declare global {
    interface Window {
        Capacitor?: any;  // Adjust as per actual Capacitor type if needed
    }
}

export interface Env {
    apiClient: ApiClient
    calendarClient: calendar.ApiClient
}

export async function getWebEnv(): Promise<Env> {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient()
    };
}

