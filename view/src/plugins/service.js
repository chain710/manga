import axios from "axios";

function convertNumberOfRect(n) {
  return Math.floor(1000 * n)
    .toString(16)
    .padStart(3, "0");
}

export class Service {
  constructor() {
    let http = axios.create({
      withCredentials: true,
      headers: { "X-Requested-With": "XMLHttpRequest" },
      delayed: process.env.VUE_APP_DELAYED_API === "true",
    });
    this.http = http;
    this.http.interceptors.request.use((config) => {
      if (config.delayed) {
        const delay = 600;
        console.warn(`delay http request ${delay}ms`);
        return new Promise((resolve) => setTimeout(() => resolve(config), 600));
      }
      return config;
    });
  }

  addLibrary(data) {
    return this.http.post(`apis/v1/library`, data);
  }

  listLibraries() {
    return this.http.get("apis/v1/library");
  }

  getLibrary(id) {
    return this.http.get(`apis/v1/library/${id}`);
  }

  patchLibrary(id, data) {
    return this.http.patch(`apis/v1/library/${id}`, { name: data.name });
  }

  deleteLibrary(id) {
    return this.http.delete(`apis/v1/library/${id}`);
  }

  scanLibrary(id) {
    return this.http.get(`apis/v1/library/${id}/scan`);
  }

  listBooks(option) {
    return this.http.get("apis/v1/book", { params: option });
  }

  getBook(id) {
    return this.http.get(`apis/v1/book/${id}`);
  }

  getVolume(id) {
    return this.http.get(`apis/v1/volume/${id}`);
  }

  listVolume(option) {
    return this.http.get(`apis/v1/volume`, { params: option });
  }

  updateVolumeProgress(vid, page) {
    const data = { op: "Update", volumes: [{ id: Number(vid), page: Number(page) }] };
    return this.http.post(`apis/v1/batch/volume/progress`, data);
  }

  setVolumeThumb(vid, buffer) {
    return this.http.post(`apis/v1/volume/${vid}/thumb`, buffer);
  }

  setBookThumb(bid, buffer) {
    return this.http.post(`apis/v1/book/${bid}/thumb`, buffer);
  }

  cropImage(vid, page, rect) {
    const left = convertNumberOfRect(rect.left);
    const top = convertNumberOfRect(rect.top);
    const width = convertNumberOfRect(rect.width);
    const height = convertNumberOfRect(rect.height);
    const r = `${left}${top}${width}${height}`;
    return this.http.get(`apis/v1/volume/${vid}/crop/${page}/${r}`, { responseType: "arraybuffer" });
  }

  fsListDirectory(path) {
    return this.http.get(`apis/v1/fs/listdir`, { params: { path } });
  }

  pageURL(vid, page) {
    return `apis/v1/volume/${vid}/read/${page}`;
  }
  pageThumbURL(vid, page) {
    return `apis/v1/volume/${vid}/read/${page}/thumb`;
  }
  volumeThumbURL(vid) {
    return `apis/v1/volume/${vid}/thumb`;
  }
  bookThumbURL(bid) {
    return `apis/v1/book/${bid}/thumb`;
  }
}

export default {
  install(vue) {
    vue.prototype.$service = new Service();
  },
};
