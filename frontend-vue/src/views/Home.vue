<template>
  <div>
    <form v-on:submit.prevent="getGemDependencies(gemName)">
      <div class="form-group">
        <input v-model="gemName" type="text" id="gem-input" placeholder="Enter a gem" class="form-control">
      </div>
      <div class="form-group">
        <button class="btn btn-primary">Find Gem</button>
      </div>
    </form>
    <d3-network :net-nodes="nodes" :net-links="links" :options="options">
    </d3-network>
    <loading :active="isLoading"
              :is-full-page="fullPage"/>
  </div>
</template>

<script>
import axios from 'axios';
import D3Network from 'vue-d3-network';
import Loading from 'vue-loading-overlay';
import 'vue-loading-overlay/dist/vue-loading.css';

export default {
  name: "Home",
  data() { return {
    gemName: '',
    isLoading: false,
    fullPage: true,
    nodes: [],
    links: [],
    options:
      {
        force: 1000,
        nodeSize: 10,
        nodeLabels: true,
        linkWidth:5,
      }
  } },
  methods: {
    async getGemDependencies(gemName) {
      var url = 'http://localhost:8080/gem/' + gemName
      this.isLoading = true
      axios.get(url)
      .then((response) => {
        this.nodes = JSON.parse(response.data.nodes)
        this.links = JSON.parse(response.data.links)
        this.isLoading = false
      })
      .catch((error) => {
        // handle error
        console.log("oopsie")
        console.log(error);
      })
    },
  },
  components: {
    D3Network,
    Loading
  },
};
</script>
<style src="vue-d3-network/dist/vue-d3-network.css"></style>