import axios from "axios";

export async function getResults() {
    const res = await axios.get("/api/results");
    return res.data;
}

export async function getKline(code) {
    const res = await axios.get(`/api/kline?code=${code}`);
    return res.data;
}
