import {ApiClient} from "./api_client";
import * as calendar from 'calendar';

export interface Env {
    apiClient: ApiClient,
    calendarClient: calendar.ApiClient
}