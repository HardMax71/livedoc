import { connect, MqttClient } from 'mqtt';

interface MqttConfig {
    brokerUrl: string;
    clientId: string;
    username?: string;
    password?: string;
}

export class MqttManager {
    private client: MqttClient | null = null;
    private subscriptions: Map<string, ((message: any) => void)[]> = new Map();

    async connect(config: MqttConfig): Promise<void> {
        return new Promise((resolve, reject) => {
            this.client = connect(config.brokerUrl, {
                clientId: config.clientId,
                username: config.username,
                password: config.password,
                clean: true,
            });

            this.client.on('connect', () => {
                console.log('Connected to MQTT broker');
                resolve();
            });

            this.client.on('error', (error) => {
                console.error('MQTT connection error:', error);
                reject(error);
            });

            this.client.on('message', (topic, message) => {
                const handlers = this.subscriptions.get(topic);
                if (handlers) {
                    const payload = JSON.parse(message.toString());
                    handlers.forEach(handler => handler(payload));
                }
            });
        });
    }

    subscribe(topic: string, handler: (message: any) => void): void {
        if (!this.client) {
            throw new Error('MQTT client not connected');
        }

        const handlers = this.subscriptions.get(topic) || [];
        if (handlers.length === 0) {
            this.client.subscribe(topic);
        }
        handlers.push(handler);
        this.subscriptions.set(topic, handlers);
    }

    unsubscribe(topic: string, handler: (message: any) => void): void {
        if (!this.client) return;

        const handlers = this.subscriptions.get(topic);
        if (handlers) {
            const index = handlers.indexOf(handler);
            if (index > -1) {
                handlers.splice(index, 1);
            }
            if (handlers.length === 0) {
                this.client.unsubscribe(topic);
                this.subscriptions.delete(topic);
            }
        }
    }

    publish(topic: string, message: any): void {
        if (!this.client) {
            throw new Error('MQTT client not connected');
        }

        this.client.publish(topic, JSON.stringify(message));
    }

    disconnect(): void {
        if (this.client) {
            this.client.end();
            this.client = null;
            this.subscriptions.clear();
        }
    }
}

export const mqttManager = new MqttManager();