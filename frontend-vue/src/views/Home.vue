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
  </div>
</template>

<script>
import axios from 'axios';
import D3Network from 'vue-d3-network';

export default {
  name: "Home",
  data() { return {
    gemName: '',
    gemDeps: '',
    nodes: [],
    links: [],
    options:
      {
        force: 3000,
        nodeSize: 20,
        nodeLabels: true,
        linkWidth:5
      }
  } },
  methods: {
    getGemDependencies(gemName) {
      var url = 'http://localhost:8080/gem/' + gemName
      axios.get(url)
      .then((response) => {
        // handle success
        // console.log(response);
        this.gemDeps = response.data
        this.nodes = JSON.parse(response.data.nodes)
        this.links = JSON.parse(response.data.links)
      })
      .catch((error) => {
        // handle error
        console.log("oopsie")
        console.log(error);
      })
    },
  },
  components: {
    D3Network
  },
};
</script>
<style src="vue-d3-network/dist/vue-d3-network.css"></style>