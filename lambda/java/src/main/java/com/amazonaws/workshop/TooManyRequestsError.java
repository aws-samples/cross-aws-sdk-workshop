package com.amazonaws.workshop;

import com.fasterxml.jackson.annotation.JsonProperty;

public class TooManyRequestsError {

    @JsonProperty("Code")
    public String code;
    @JsonProperty("Message")
    public String message;

    public TooManyRequestsError(String message) {
        this.code = "TooManyRequestsError";
        this.message = String.format("TooManyRequestsError: %s", message);
    }

    public String getCode() {
        return code;
    }

    public void setCode(String code) {
        this.code = code;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }
}
