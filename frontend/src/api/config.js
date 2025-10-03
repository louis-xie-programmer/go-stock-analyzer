import axios from "axios";

export async function getConfig() {
    const res = await axios.get("/api/config");
    return res.data;
}

export async function updateConfig(cfg) {
    await axios.post("/api/config", cfg);
}
