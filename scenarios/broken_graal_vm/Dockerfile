FROM ghcr.io/graalvm/graalvm-ce:ol7-java17-22.3.0-b2

WORKDIR /app


COPY ./scenarios/ddprof_graal_vm/ ./
RUN gu install native-image

RUN javac DummyApp.java && jar cfm DummyApp.jar META-INF/MANIFEST.MF DummyApp.class

# RUN native-image -H:-DeleteLocalSymbols -jar DummyApp.jar
# Frame pointers can improve unwinding performance
# RUN native-image -H:+AddAllCharsets -H:+PreserveFramePointer -H:-DeleteLocalSymbols -jar DummyApp.jar

# This is to emit dwarf (instead of frame pointers)
RUN native-image -g -H:-DeleteLocalSymbols -jar DummyApp.jar

RUN yum update -y && yum install -y \
    xz jq wget

ADD ./profilers/ddprof/install_profiler.sh .
RUN ./install_profiler.sh /usr/local/bin

ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_NATIVE_LOG_MODE="/app/data/ddprof_log"
# If upload period is > EXECUTION_TIME_SEC, it can cause issues when shutting down containers 
ENV DD_PROFILING_UPLOAD_PERIOD="10"

CMD ddprof -l notice /app/DummyApp 11
