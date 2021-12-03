package com.amazonaws.workshop;

import com.fasterxml.jackson.databind.ObjectMapper;
import software.amazon.awssdk.enhanced.dynamodb.TableSchema;

import java.io.InputStream;
import java.io.OutputStream;

import static software.amazon.awssdk.enhanced.dynamodb.mapper.StaticAttributeTags.primaryPartitionKey;

public class WorkshopUtils {

    static final TableSchema<PodcastEpisode> GET_PODCAST_TABLE_SCHEMA =
            TableSchema.builder(PodcastEpisode.class)
                    .newItemSupplier(PodcastEpisode::new)
                    .addAttribute(String.class, a -> a.name("id")
                            .getter(PodcastEpisode::getId)
                            .setter(PodcastEpisode::setId)
                            .tags(primaryPartitionKey()))
                    .addAttribute(String.class, a -> a.name("title")
                            .getter(PodcastEpisode::getTitle)
                            .setter(PodcastEpisode::setTitle))
                    .addAttribute(String.class, a -> a.name("description")
                            .getter(PodcastEpisode::getDescription)
                            .setter(PodcastEpisode::setDescription))
                    .addAttribute(String.class, a -> a.name("podcast")
                            .getter(PodcastEpisode::getPodcast)
                            .setter(PodcastEpisode::setPodcast))
                    .addAttribute(String.class, a -> a.name("status")
                            .getter(PodcastEpisode::getStatus)
                            .setter(PodcastEpisode::setStatus))
                    .build();

    static final TableSchema<PodcastEpisode> LIST_PODCAST_TABLE_SCHEMA =
            TableSchema.builder(PodcastEpisode.class)
                    .newItemSupplier(PodcastEpisode::new)
                    .addAttribute(String.class, a -> a.name("id")
                            .getter(PodcastEpisode::getId)
                            .setter(PodcastEpisode::setId)
                            .tags(primaryPartitionKey()))
                    .addAttribute(String.class, a -> a.name("title")
                            .getter(PodcastEpisode::getTitle)
                            .setter(PodcastEpisode::setTitle))
                    .addAttribute(String.class, a -> a.name("podcast")
                            .getter(PodcastEpisode::getPodcast)
                            .setter(PodcastEpisode::setPodcast))
                    .build();

    static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    public static <T> T readValue(InputStream i, Class<T> clazz) {
        try {
            return OBJECT_MAPPER.readValue(i, clazz);
        } catch (Exception e) {
            System.out.println(e);
            throw new RuntimeException(e);
        }
    }

    public static String writeValue(Object e) {
        try {
            return OBJECT_MAPPER.writeValueAsString(e);
        } catch (Exception ex) {
            System.out.println(ex);
            throw new RuntimeException(ex);
        }
    }

    public static void writeOutputStream(OutputStream out, Object o) {
        try {
            OBJECT_MAPPER.writeValue(out, o);
        } catch (Exception e) {
            System.out.println(e);
            throw new RuntimeException("Unable to write to output stream!");
        }
    }
}
