import {ApiClient} from "./api_client";
import * as calendar from 'calendar';
import {Connector} from "rtgraph";

export interface Env {
    apiClient: ApiClient,
    calendarClient: calendar.ApiClient
    connector?: Connector,

    maybeStartFrontendBLE(): Promise<void>
}