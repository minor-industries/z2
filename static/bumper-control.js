import { ReadVariables, UpdateVariables } from './api.js';
class BumperControl {
    constructor(containerId, label, variableName, increment, fixed = null) {
        const containerElement = document.getElementById(containerId);
        if (!containerElement) {
            throw new Error(`Container with id ${containerId} not found`);
        }
        this.container = containerElement;
        this.label = label;
        this.variableName = variableName;
        this.increment = increment;
        this.speed = 0; // default speed
        this.fixed = fixed;
        this.createControl();
        this.fetchInitialSpeed();
    }
    createControl() {
        const controlHTML = `
            <div class="pure-u-1 pure-u-md-1-2 pure-u-lg-1-3 pure-u-xl-1-4">
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
        const lastElement = this.container.lastElementChild;
        this.display = lastElement.querySelector('.bumper-display');
        const incrementButton = lastElement.querySelector('.bumper-buttons button:nth-child(1)');
        const decrementButton = lastElement.querySelector('.bumper-buttons button:nth-child(2)');
        // Bind event handlers
        incrementButton.addEventListener('click', () => this.changeSpeed(this.increment));
        decrementButton.addEventListener('click', () => this.changeSpeed(-this.increment));
    }
    async fetchInitialSpeed() {
        try {
            const req = {
                variables: [this.variableName]
            };
            const resp = await ReadVariables(req);
            const variable = resp.variables.find(v => v.name === this.variableName);
            if (variable && variable.present) {
                this.speed = variable.value;
                this.updateValue();
            }
        }
        catch (error) {
            console.error('Error fetching initial speed:', error);
        }
    }
    async changeSpeed(delta) {
        this.speed += delta;
        this.updateValue();
        await this.updateSpeedBackend();
    }
    updateValue() {
        if (this.fixed !== null) {
            this.display.value = Number(this.speed).toFixed(this.fixed);
        }
        else {
            this.display.value = this.speed.toString();
        }
    }
    async updateSpeedBackend() {
        try {
            const req = {
                variables: [{
                        name: this.variableName,
                        value: this.speed,
                        present: true
                    }]
            };
            await UpdateVariables(req);
            console.log('Speed updated successfully');
        }
        catch (error) {
            console.error('Error updating speed:', error);
        }
    }
}
export default BumperControl;
