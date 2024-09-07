import {calendar, DefaultApiClient} from "/dist/z2-bundle.js";

export async function getEnv() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient()
    };
}
