import { ReadVariablesReq, ReadVariablesResp, UpdateVariablesReq, UpdateVariables, ReadVariables } from './api.js';

class BumperControl {
    private container: HTMLElement;
    private label: string;
    private variableName: string;
    private increment: number;
    private decrement: number;
    private speed: number;
    private display!: HTMLInputElement;

    constructor(containerId: string, label: string, variableName: string, increment: number = 1, decrement: number = 1) {
        const containerElement = document.getElementById(containerId);
        if (!containerElement) {
            throw new Error(`Container with id ${containerId} not found`);
        }
        this.container = containerElement;
        this.label = label;
        this.variableName = variableName;
        this.increment = increment;
        this.decrement = decrement;
        this.speed = 0; // default speed

        this.createControl();
        this.fetchInitialSpeed();
    }

    private createControl(): void {
        const controlHTML = `
            <div class="pure-u-1">
                <div class="bumper-container">
                    <label style="margin-right: 10px;">${this.label}</label>
                    <input type="text" class="pure-input-1 bumper-display" value="${this.speed}" readonly>
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
        incrementButton.addEventListener('click', () => this.changeSpeed(this.increment));
        decrementButton.addEventListener('click', () => this.changeSpeed(-this.decrement));
    }

    private async fetchInitialSpeed(): Promise<void> {
        try {
            const req: ReadVariablesReq = {
                variables: [this.variableName]
            };

            const resp: ReadVariablesResp = await ReadVariables(req);
            const variable = resp.variables.find(v => v.name === this.variableName);
            if (variable && variable.present) {
                this.speed = variable.value;
                this.display.value = this.speed.toString();
            }
        } catch (error) {
            console.error('Error fetching initial speed:', error);
        }
    }

    public async changeSpeed(delta: number): Promise<void> {
        this.speed += delta;
        this.display.value = this.speed.toString();
        await this.updateSpeedBackend();
    }

    private async updateSpeedBackend(): Promise<void> {
        try {
            const req: UpdateVariablesReq = {
                variables: [{
                    name: this.variableName,
                    value: this.speed,
                    present: true
                }]
            };

            await UpdateVariables(req);
            console.log('Speed updated successfully');
        } catch (error) {
            console.error('Error updating speed:', error);
        }
    }
}

export default BumperControl;
