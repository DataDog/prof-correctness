FROM adoptopenjdk/openjdk11:alpine

RUN apk update && apk add tar zip unzip curl moreutils wget hexdump linux-headers bash make g++ git maven cppcheck jq
RUN wget -nv https://search.maven.org/remotecontent?filepath=com/datadoghq/ap-tools/2.8.3-DD-20221216/ap-tools-2.8.3-DD-20221216.jar -O ap-tools.jar
RUN wget -nv https://github.com/renaissance-benchmarks/renaissance/releases/download/v0.14.1/renaissance-mit-0.14.1.jar -O renaissance.jar
RUN unzip ap-tools.jar
RUN echo "java -agentpath:./native-libs/linux-musl-x64/libasyncProfiler.so -jar renaissance.jar -r 2 akka-uct" > run.sh && chmod a+x run.sh

CMD ./run.sh
