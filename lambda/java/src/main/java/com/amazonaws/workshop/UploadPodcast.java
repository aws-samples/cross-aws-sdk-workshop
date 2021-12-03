package com.amazonaws.workshop;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestStreamHandler;
import software.amazon.awssdk.core.async.AsyncRequestBody;
import software.amazon.awssdk.core.internal.util.Mimetype;
import software.amazon.awssdk.services.dynamodb.model.ProvisionedThroughputExceededException;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;
import software.amazon.awssdk.transfer.s3.S3TransferManager;
import software.amazon.awssdk.utils.IoUtils;
import software.amazon.awssdk.utils.StringUtils;

import java.io.InputStream;
import java.io.OutputStream;
import java.net.URI;
import java.net.URLConnection;
import java.nio.file.Paths;

public class UploadPodcast implements RequestStreamHandler {

    private static final String S3_PODCAST_PREFIX = System.getenv("AWS_SDK_WORKSHOP_PODCAST_DATA_KEY_PREFIX");
    private static final String PODCAST_BUCKET = System.getenv("AWS_SDK_WORKSHOP_PODCAST_DATA_BUCKET_NAME");

    private final S3TransferManager s3;

    public UploadPodcast() {
        this.s3 = S3TransferManager.create();
    }

    @Override
    public void handleRequest(InputStream input, OutputStream outputStream, Context context) {
        try {
            Episode episode = WorkshopUtils.readValue(input, Episode.class);
            PodcastEpisode podcastEpisode = episode.getEpisode();

            URLConnection podcastUrl = downloadPodcast(podcastEpisode);

            String objectKey = String.format("%s%s/raw-media", S3_PODCAST_PREFIX, podcastEpisode.getId());
            String contentType = determineContentType(podcastEpisode, podcastUrl);
            episode.getEpisode().setMediaContentType(contentType);
            episode.getEpisode().setMediaKey(objectKey);
            
            System.out.println(String.format("Uploading podcast to S3 URI: %s/%s with content type %s", PODCAST_BUCKET, objectKey, contentType));

            byte[] podcast = IoUtils.toByteArray(podcastUrl.getInputStream());

            uploadMedia(podcast, PODCAST_BUCKET, objectKey, contentType);

            WorkshopUtils.writeOutputStream(outputStream, episode);
        } catch (ProvisionedThroughputExceededException e) {
            System.out.println(WorkshopUtils.writeValue(new TooManyRequestsError("Please slow down request rate")));
        } catch (RuntimeException e) {
            System.out.println(e);
        } catch (Exception e) {
            System.out.println(new RuntimeException("Unable to upload podcast!", e));
        }
    }

    private void uploadMedia(byte[] podcast, String bucket, String key, String contentType) {
        AsyncRequestBody asyncRequestBody = AsyncRequestBody.fromBytes(podcast);
        PutObjectRequest request = PutObjectRequest.builder().bucket(PODCAST_BUCKET).key(key).contentType(contentType).build();
        s3.upload(r -> r.requestBody(asyncRequestBody).putObjectRequest(request)).completionFuture().join();
    }

    private URLConnection downloadPodcast(PodcastEpisode podcastEpisode) throws Exception {
        System.out.println(String.format("Downloading podcast from %s", podcastEpisode.getMediaUrl()));
        return URI.create(podcastEpisode.getMediaUrl()).toURL().openConnection();
    }

    private String determineContentType(PodcastEpisode podcastEpisode, URLConnection connection) throws Exception {
        if (StringUtils.isNotBlank(podcastEpisode.getMediaContentType())) {
            return podcastEpisode.getMediaContentType();
        }

        if (StringUtils.isNotBlank(connection.getContentType())) {
            return connection.getContentType();
        }

        String mimeTypeFromUrl = Mimetype.getInstance().getMimetype(Paths.get(connection.getURL().toURI()));

        if (StringUtils.isNotBlank(mimeTypeFromUrl)) {
            return mimeTypeFromUrl;
        }

        throw new RuntimeException("Unable to determine content-type!");
    }
}
