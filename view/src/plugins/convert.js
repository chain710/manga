function readPercent(page, count) {
  if (count == 0) {
    return 0;
  }

  return Math.ceil((100 * page) / count);
}
export default {
  install(vue) {
    vue.prototype.$convertBook = function (book) {
      if (!book) {
        return null;
      }
      return {
        id: book.id,
        title: { text: book.name, to: { name: "book", params: { bookID: book.id } } },
        thumb: `/apis/v1/book/${book.id}/thumb`,
        subTitle: { text: book.writer },
        bottomText: this.$t("global.total_volumes", { count: book.volume }),
        unread: book.progress == undefined, // boolean show unread triangle
        readPercent: 0, // read status 0-100
        readable: false, // boolean false or string
      };
    };

    vue.prototype.$convertBooks = function (books) {
      let ret = [];
      for (let book of books) {
        ret.push(this.$convertBook(book));
      }
      return ret;
    };

    vue.prototype.$convertVolume = function (volume, reading) {
      if (!volume) {
        return null;
      }

      let readable = {
        name: "read",
        params: { bookID: volume.book_id, volumeID: volume.id },
      };

      let rp = 0;
      if (volume.progress) {
        rp = readPercent(volume.progress.page, volume.page_count);
        readable["query"] = { page: volume.progress.page };
      }

      const result = {
        id: volume.id,
        title: { text: volume.title },
        thumb: `/apis/v1/volume/${volume.id}/thumb`,
        subTitle: { text: volume.book_name },
        bottomText: null,
        readPercent: rp,
        unread: rp == 0,
        readable,
      };
      if (volume.book_name) {
        result.title = { text: volume.book_name, to: { name: "book", params: { bookID: volume.book_id } } };
        result.subTitle = { text: volume.title };
      }

      if (reading) {
        result.bottomText = volume.volume
          ? this.$t("global.reading_volume", { num: volume.volume, page: volume.page_count })
          : this.$t("global.reading_extra", { page: volume.page_count });
      } else {
        result.bottomText = volume.volume
          ? this.$t("global.volume_at", { num: volume.volume, page: volume.page_count })
          : this.$t("global.extra_at", { page: volume.page_count });
      }

      return result;
    };

    vue.prototype.$convertBookToReadingVolume = function (book) {
      if (!book.progress) {
        console.error(`book ${book.id} progress undefined`);
        return null;
      }
      const vid = book.progress.volume_id;
      const rp = readPercent(book.progress.page, book.progress.page_count);
      console.debug(`reading vol`, book);
      return {
        id: vid,
        title: { text: book.progress.title, to: { name: "book", params: { bookID: book.id } } },
        thumb: `/apis/v1/volume/${vid}/thumb`,
        subTitle: { text: "" },
        bottomText: this.$t("global.reading_volume", {
          num: book.progress.volume,
          page: book.progress.page,
        }),
        readPercent: rp,
        unread: rp == 0,
        readable: {
          name: "read",
          params: { bookID: book.progress.book_id, volumeID: vid },
          query: { page: book.progress.page },
        },
      };
    };

    vue.prototype.$convertVolumes = function (volumes, reading) {
      let ret = [];
      for (let vol of volumes) {
        ret.push(this.$convertVolume(vol, reading));
      }
      return ret;
    };

    vue.prototype.$convertBooksToReadingVolumes = function (books) {
      return books.map((x) => this.$convertBookToReadingVolume(x));
    };
  },
};
