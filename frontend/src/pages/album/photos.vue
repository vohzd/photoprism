<template>
    <div class="p-page p-page-album-photos" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <p-album-toolbar :album="model" :settings="settings" :filter="filter" :filter-change="updateQuery"
                              :refresh="refresh"></p-album-toolbar>

        <v-container fluid class="pa-4" v-if="loading">
            <v-progress-linear color="secondary-dark"  :indeterminate="true"></v-progress-linear>
        </v-container>
        <v-container fluid class="pa-0" v-else>
            <p-scroll-top></p-scroll-top>

            <p-photo-clipboard :refresh="refresh"
                               :selection="selection"
                               :album="model" context="album"></p-photo-clipboard>

            <p-photo-mosaic v-if="settings.view === 'mosaic'"
                            :photos="results"
                            :selection="selection"
                            :album="model"
                            :open-photo="openPhoto"></p-photo-mosaic>
            <p-photo-list v-else-if="settings.view === 'list'"
                          :photos="results"
                          :selection="selection"
                          :album="model"
                          :open-photo="openPhoto"
                          :open-location="openLocation"></p-photo-list>
            <p-photo-details v-else-if="settings.view === 'details'"
                             :photos="results"
                             :selection="selection"
                             :album="model"
                             :open-photo="openPhoto"
                             :open-location="openLocation"></p-photo-details>
            <p-photo-tiles v-else :photos="results"
                           :selection="selection"
                           :album="model"
                           :open-photo="openPhoto"></p-photo-tiles>
        </v-container>
    </div>
</template>

<script>
    import Photo from "model/photo";
    import Album from "model/album";

    export default {
        name: 'p-page-album-photos',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.uuid = this.$route.params.uuid;
                this.filter.q = query['q'] ? query['q'] : '';
                this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
                this.filter.country = query['country'] ? query['country'] : '';
                this.settings.view = this.viewType();
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.findAlbum();
                this.search();
            }
        },
        data() {
            const uuid = this.$route.params.uuid;
            const query = this.$route.query;
            const routeName = this.$route.name;
            const order = query['order'] ? query['order'] : 'oldest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = this.viewType();
            const filter = {country: country, camera: camera, order: order, q: q};
            const settings = {view: view};

            return {
                model: new Album({AlbumName:""}),
                uuid: uuid,
                results: [],
                scrollDisabled: true,
                pageSize: 60,
                offset: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                loading: true
            };
        },
        methods: {
            viewType() {
                let queryParam = this.$route.query['view'];
                let defaultType = window.localStorage.getItem("photo_view_type");
                let storedType = window.localStorage.getItem("album_view_type");

                if (queryParam) {
                    window.localStorage.setItem("album_view_type", queryParam);
                    return queryParam;
                } else if (storedType) {
                    return storedType;
                } else if (defaultType) {
                    return defaultType;
                } else if (window.innerWidth < 960) {
                    return 'mosaic';
                } else if (window.innerWidth > 1600) {
                    return 'details';
                }

                return 'tiles';
            },
            openLocation(index) {
                const photo = this.results[index];

                if (photo.PhotoLat && photo.PhotoLng) {
                    this.$router.push({name: "places", query: {lat: String(photo.PhotoLat), lng: String(photo.PhotoLng)}});
                } else if (photo.LocCity) {
                    this.$router.push({name: "places", query: {q: photo.LocCity}});
                }
            },
            openPhoto(index) {
                this.$viewer.show(this.results, index)
            },
            loadMore() {
                if (this.scrollDisabled) return;

                this.scrollDisabled = true;

                this.offset += this.pageSize;

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                    album: this.uuid,
                };

                Object.assign(params, this.lastFilter);

                Photo.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info(this.$gettext('All ') + this.results.length + this.$gettext(' photos loaded'));
                    }
                });
            },
            updateQuery() {
                const query = {
                    view: this.settings.view
                };

                Object.assign(query, this.filter);

                for (let key in query) {
                    if (query[key] === undefined || !query[key]) {
                        delete query[key];
                    }
                }

                if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
                    return
                }

                this.$router.replace({query: query});
            },
            searchParams() {
                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                    album: this.uuid,
                };

                Object.assign(params, this.filter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return params;
            },
            refresh() {
                this.lastFilter = {};
                const pageSize = this.pageSize;
                this.pageSize = this.offset + pageSize;
                this.search();
                this.offset = this.pageSize;
                this.pageSize = pageSize;
            },
            search() {
                this.scrollDisabled = true;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
                    this.$nextTick(() => this.$emit("scrollRefresh"));
                    return;
                }

                Object.assign(this.lastFilter, this.filter);

                this.offset = 0;
                this.loading = true;

                const params = this.searchParams();

                Photo.search(params).then(response => {
                    this.loading = false;
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        if (!this.results.length) {
                            this.$notify.warning(this.$gettext("No photos found"));
                        } else if (this.results.length === 1) {
                            this.$notify.info("One photo found");
                        } else {
                            this.$notify.info(this.results.length + this.$gettext(" photos found"));
                        }
                    } else {
                        this.$notify.info(this.$gettext('More than 50 photos found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).catch(() => this.loading = false);
            },
            findAlbum() {
                this.model.find(this.uuid).then(m => {
                    this.model = m;
                    window.document.title = `PhotoPrism: ${this.model.AlbumName}`;
                });
            },
        },
        created() {
            this.findAlbum();
            this.search();
        },
    };
</script>
