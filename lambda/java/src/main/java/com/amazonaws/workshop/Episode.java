package com.amazonaws.workshop;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Episode {

    @JsonProperty("episode")
    private PodcastEpisode episode;

    public Episode() {
    }

    @JsonProperty("episode")
    public PodcastEpisode getEpisode() {
        return episode;
    }

    @JsonProperty("episode")
    public void setEpisode(PodcastEpisode episode) {
        this.episode = episode;
    }
}
