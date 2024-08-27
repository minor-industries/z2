import {ApiClient} from "./api_client";

interface BumperControlConfig {
    containerId: string;
    label: string;
    variableName: string;
    increment: number;
    defaultValue: number;
    fixed?: number;
    maxValue?: number;
    minValue?: number;
}

export class BumperControl {
    private container: HTMLElement;
    private readonly label: string;
    private readonly variableName: string;
    private readonly increment: number;
    private value: number; // renamed from speed
    private readonly fixed?: number;
    private readonly maxValue?: number;
    private readonly minValue?: number;
    private readonly defaultValue: number;
    private display!: HTMLInputElement;
    private apiClient: ApiClient;

    constructor(apiClient: ApiClient, config: BumperControlConfig) {
        this.apiClient = apiClient;
        const containerElement = document.getElementById(config.containerId);
        if (!containerElement) {
            throw new Error(`Container with id ${config.containerId} not found`);
        }
        this.container = containerElement;
        this.label = config.label;
        this.variableName = config.variableName;
        this.increment = config.increment;
        this.value = 0; // default value
        this.fixed = config.fixed;
        this.maxValue = config.maxValue;
        this.minValue = config.minValue;
        this.defaultValue = config.defaultValue;

        this.createControl();
        this.fetchInitialValue();
    }

    private createControl(): void {
        const controlHTML = `
            <div class="pure-u-1 pure-u-md-1-2 pure-u-lg-1-3 pure-u-xl-1-4">
                <div class="bumper-container">
                    <label style="margin-right: 10px;">${this.label}</label>
                    <input type="text" class="pure-input-1 bumper-display" value="${this.value}" readonly>
                    <div class="bumper-buttons">
                        <button class="pure-button pure-button-primary bumper-button">▲</button>
                        <button class="pure-button pure-button-primary bumper-button">▼</button>
                    </div>
                </div>
            </div>
        `;
        this.container.insertAdjacentHTML('beforeend', controlHTML);
        const lastElement = this.container.lastElementChild as HTMLElement;
        this.display = lastElement.querySelector('.bumper-display') as HTMLInputElement;
        const incrementButton = lastElement.querySelector('.bumper-buttons button:nth-child(1)') as HTMLButtonElement;
        const decrementButton = lastElement.querySelector('.bumper-buttons button:nth-child(2)') as HTMLButtonElement;

        // Bind event handlers
        incrementButton.addEventListener('click', () => this.changeValue(this.increment));
        decrementButton.addEventListener('click', () => this.changeValue(-this.increment));
    }

    private async fetchInitialValue(): Promise<void> {
        try {
            const resp = await this.apiClient.readVariables({
                variables: [this.variableName]
            });
            const variable = resp.variables.find(v => v.name === this.variableName);
            if (variable && variable.present) {
                this.value = variable.value;
            } else {
                this.value = this.defaultValue;
                await this.updateValueBackend();
            }
            this.updateDisplay();
        } catch (error) {
            console.error('Error fetching initial value:', error);
        }
    }

    public async changeValue(delta: number): Promise<void> {
        const newValue = this.value + delta;
        await this.setValue(newValue);
    }

    public async setValue(newValue: number): Promise<void> {
        this.value = this.validateValue(newValue);
        this.updateDisplay();
        await this.updateValueBackend();
    }

    public getVariableName(): string {
        return this.variableName;
    }

    private validateValue(value: number): number {
        if (this.maxValue !== undefined && value > this.maxValue) {
            return this.maxValue;
        }
        if (this.minValue !== undefined && value < this.minValue) {
            return this.minValue;
        }
        return value;
    }

    private updateDisplay() {
        if (this.fixed !== undefined) {
            this.display.value = Number(this.value).toFixed(this.fixed);
        } else {
            this.display.value = this.value.toString();
        }
    }

    private async updateValueBackend(): Promise<void> {
        try {
            await this.apiClient.updateVariables({
                variables: [{
                    name: this.variableName,
                    value: this.value,
                    present: true
                }]
            });
            console.log('Value updated successfully');
        } catch (error) {
            console.error('Error updating value:', error);
        }
    }
}
