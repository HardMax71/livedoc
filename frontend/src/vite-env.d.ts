/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_API_URL: string
    readonly VITE_MQTT_BROKER: string
    readonly VITE_MQTT_PORT: string
    readonly VITE_MQTT_PATH: string
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}